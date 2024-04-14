package factory

import (
	"context"
	"fmt"
	"github.com/expgo/generic"
	"github.com/expgo/sync"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"reflect"
	"time"
)

var _context = &factoryContext{
	exprEnvInitOnce: sync.NewOnce(),
}

type Getter func(ctx context.Context) any

type factoryContext struct {
	defaultMustBuilderCache generic.Map[reflect.Type, *contextCachedItem] // package:name -> must builder
	namedMustBuilderCache   generic.Map[string, *contextCachedItem]       // name -> must builder
	wiringCache             generic.Map[reflect.Type, bool]
	exprEnv                 *exprEnv
	exprEnvInitOnce         sync.Once
}

type contextCachedItem struct {
	_type  reflect.Type
	getter Getter
}

type exprEnv struct {
	env  map[string]any
	lock sync.Mutex
	ctx  context.Context
}

func (c *exprEnv) Visit(node *ast.Node) {
	if s, ok := (*node).(*ast.IdentifierNode); ok {
		_, ok = c.getValue(s.String())
		if !ok {
			value := _context.getByName(c.ctx, s.String())
			c.setValue(s.String(), value)
		}
	}
}

func (c *exprEnv) getValue(name string) (any, bool) {
	value, ok := c.env[name]
	return value, ok
}

func (c *exprEnv) setValue(name string, value any) {
	c.env[name] = value
}

func Find[T any]() *T {
	return FindTimeout[T](Opts.DefaultTimeout)
}

func FindTimeout[T any](timeout time.Duration) *T {

	vt := reflect.TypeOf((*T)(nil))

	result := _context.getByType(getTimeoutContext(timeout), vt)

	resultType := reflect.TypeOf(result)
	if resultType.Kind() == reflect.Ptr && resultType.ConvertibleTo(vt) {
		return result.(*T)
	}

	// panic
	panic(fmt.Errorf("Invalid type: need %v, get %v", vt, resultType))
}

func FindByName[T any](name string) *T {
	return FindByNameTimeout[T](name, Opts.DefaultTimeout)
}

func FindByNameTimeout[T any](name string, timeout time.Duration) *T {
	vt := reflect.TypeOf((*T)(nil))

	result := _context.getByName(getTimeoutContext(timeout), name)

	resultType := reflect.TypeOf(result)
	if resultType.Kind() == reflect.Ptr && resultType.ConvertibleTo(vt) {
		return result.(*T)
	}

	// panic
	panic(fmt.Errorf("Invalid type: need %v, get %v", vt, resultType))
}

func Range[T any](rangeFunc func(any) bool) {
	RangeTimeout[T](rangeFunc, Opts.DefaultTimeout)
}

func RangeTimeout[T any](rangeFunc func(any) bool, timeout time.Duration) {
	rangeContext[T](rangeFunc, getTimeoutContext(timeout))
}

func rangeContext[T any](rangeFunc func(any) bool, ctx context.Context) {
	vt := reflect.TypeOf((*T)(nil))

	if vt.Elem().Kind() == reflect.Interface {
		vt = vt.Elem()
	} else if vt.Elem().Kind() != reflect.Struct {
		panic("Range only range type and interface")
	}

	_context.defaultMustBuilderCache.Range(func(k reflect.Type, v *contextCachedItem) bool {
		if k.ConvertibleTo(vt) {
			return rangeFunc(v.getter(ctx))
		}
		return true
	})
}

func RangeStruct[T any](structFunc func(*T) bool) {
	vt := reflect.TypeOf((*T)(nil))

	if vt.Elem().Kind() == reflect.Struct {
		Range[T](func(v any) bool {
			return structFunc(v.(*T))
		})
	} else {
		panic("Range type only range struct type")
	}
}

func FindStructs[T any]() (result []*T) {
	vt := reflect.TypeOf((*T)(nil))

	if vt.Elem().Kind() == reflect.Struct {
		Range[T](func(v any) bool {
			result = append(result, v.(*T))
			return true
		})
	} else {
		panic("Range type only range struct type")
	}

	return
}

func RangeInterface[T any](interfaceFunc func(T) bool) {
	vt := reflect.TypeOf((*T)(nil))

	if vt.Elem().Kind() == reflect.Interface {
		Range[T](func(v any) bool {
			return interfaceFunc(v.(T))
		})
	} else {
		panic("Range inf only range interface type")
	}
}

func FindInterfaces[T any]() (result []T) {
	vt := reflect.TypeOf((*T)(nil))

	if vt.Elem().Kind() == reflect.Interface {
		Range[T](func(v any) bool {
			result = append(result, v.(T))
			return true
		})
	} else {
		panic("Range inf only range interface type")
	}

	return
}

func (c *factoryContext) getByNameOrType(ctx context.Context, name string, vt reflect.Type) any {
	mb, ok := c.namedMustBuilderCache.Load(name)

	if ok {
		result := mb.getter(ctx)
		rt := reflect.TypeOf(result)
		if vt.ConvertibleTo(rt) {
			return result
		}
	}

	return c.getByType(ctx, vt)
}

func (c *factoryContext) getByType(ctx context.Context, vt reflect.Type) any {
	mb, ok := c.defaultMustBuilderCache.Load(vt)

	if ok {
		return mb.getter(ctx)
	}

	if vt.Kind() == reflect.Interface {
		// 需求是接口才使用下面方法找寻
		convertibleList := c.defaultMustBuilderCache.Filter(func(k reflect.Type, v *contextCachedItem) bool {
			return k.ConvertibleTo(vt)
		})

		convertibleListSize := convertibleList.Size()

		if convertibleListSize > 1 {
			panic(fmt.Errorf("Multiple default builders found for type: %v, please use named singleton", vt))
		}

		if convertibleListSize == 1 {
			convertibleList.Range(func(k reflect.Type, v *contextCachedItem) bool {
				mb = v
				ok = true
				return false
			})

			if ok {
				return mb.getter(ctx)
			}
		}
	}

	svt := vt
	if svt.Kind() == reflect.Ptr {
		svt = svt.Elem()
	}

	panic(fmt.Errorf("use type to get Getter, %s:%s not found.", svt.PkgPath(), svt.Name()))

}

func (c *factoryContext) setByType(vt reflect.Type, builder Getter) {
	_, getOk := c.defaultMustBuilderCache.LoadOrStore(vt, &contextCachedItem{_type: vt, getter: builder})
	if getOk {
		panic(fmt.Errorf("Default builder allready exist: %s", vt.String()))
	}
}

func (c *factoryContext) getByName(ctx context.Context, name string) any {
	mb, ok := c.namedMustBuilderCache.Load(name)

	if ok {
		return mb.getter(ctx)
	}

	panic(fmt.Errorf("Named builder %s not found.", name))
}

func (c *factoryContext) setByName(name string, vt reflect.Type, builder Getter) {
	_, getOk := c.namedMustBuilderCache.LoadOrStore(name, &contextCachedItem{_type: vt, getter: builder})
	if getOk {
		panic(fmt.Errorf("Named builder allready exist: %s", name))
	}
}

func (c *factoryContext) wiring(vt reflect.Type) {
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}

	_, ok := c.wiringCache.LoadOrStore(vt, true)
	if ok {
		panic(fmt.Errorf("%s:%s is wiring, possible circular reference exists.", vt.PkgPath(), vt.Name()))
	}
}

func (c *factoryContext) wired(vt reflect.Type) {
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}

	c.wiringCache.Delete(vt)
}

func (c *factoryContext) evalExpr(ctx context.Context, code string) (any, error) {
	err := c.exprEnvInitOnce.Do(func() error {
		c.exprEnv = &exprEnv{
			env:  make(map[string]any),
			lock: sync.NewMutex(),
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	tree, _ := parser.Parse(code)

	c.exprEnv.lock.Lock()
	defer c.exprEnv.lock.Unlock()
	c.exprEnv.ctx = ctx
	ast.Walk(&tree.Node, c.exprEnv)

	return expr.Eval(code, c.exprEnv.env)
}

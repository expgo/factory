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
	exprEnvMap:  make(map[string]any),
	exprEnvLock: sync.NewRWMutex(),
}

type Getter func(ctx context.Context) any

type factoryContext struct {
	typedMap    generic.Map[reflect.Type, *contextCachedItem] // package:name -> must builder
	namedMap    generic.Map[string, *contextCachedItem]       // name -> must builder
	wiringCache generic.Map[reflect.Type, bool]
	exprEnvMap  map[string]any
	exprEnvLock sync.RWMutex
}

type contextCachedItem struct {
	_type  reflect.Type
	getter Getter
}

type exprContext struct {
	ctx context.Context
}

func (c *exprContext) Visit(node *ast.Node) {
	if s, ok := (*node).(*ast.IdentifierNode); ok {
		_, ok = c.getValue(s.String())
		if !ok {
			value := _context.getByNamePanic(c.ctx, s.String(), nil)
			c.setValue(s.String(), value)
		}
	}
}

func (c *exprContext) getValue(name string) (any, bool) {
	_context.exprEnvLock.RLock()
	defer _context.exprEnvLock.RUnlock()

	value, ok := _context.exprEnvMap[name]
	return value, ok
}

func (c *exprContext) setValue(name string, value any) {
	_context.exprEnvLock.Lock()
	defer _context.exprEnvLock.Unlock()

	_context.exprEnvMap[name] = value
}

func Find[T any]() *T {
	return FindTimeout[T](Opts.Timeout)
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
	return FindByNameTimeout[T](name, Opts.Timeout)
}

func FindByNameTimeout[T any](name string, timeout time.Duration) *T {
	vt := reflect.TypeOf((*T)(nil))

	result := _context.getByNamePanic(getTimeoutContext(timeout), name, vt)

	resultType := reflect.TypeOf(result)
	if resultType.Kind() == reflect.Ptr && resultType.ConvertibleTo(vt) {
		return result.(*T)
	}

	// panic
	panic(fmt.Errorf("Invalid type: need %v, get %v", vt, resultType))
}

func Range[T any](rangeFunc func(any) bool) {
	RangeTimeout[T](rangeFunc, Opts.Timeout)
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

	_context.typedMap.Range(func(k reflect.Type, v *contextCachedItem) bool {
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
	if ret, err := c.getByName(ctx, name, vt); err == nil {
		return ret
	}

	return c.getByType(ctx, vt)
}

func (c *factoryContext) getByType(ctx context.Context, vt reflect.Type) any {
	mb, ok := c.typedMap.Load(vt)

	if ok {
		ctx = pushGetter(ctx, mb)
		defer popGetter(ctx)

		return mb.getter(ctx)
	}

	if vt.Kind() == reflect.Interface {
		// 需求是接口才使用下面方法找寻
		convertibleList := c.typedMap.Filter(func(k reflect.Type, v *contextCachedItem) bool {
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
				ctx = pushGetter(ctx, mb)
				defer popGetter(ctx)

				return mb.getter(ctx)
			}
		}
	}

	svt := vt
	if svt.Kind() == reflect.Ptr {
		svt = svt.Elem()
	}

	panic(fmt.Errorf("use type to get Getter, %s:%s not found", svt.PkgPath(), svt.Name()))

}

func (c *factoryContext) setByType(vt reflect.Type, cci *contextCachedItem) {
	_, getOk := c.typedMap.LoadOrStore(vt, cci)
	if getOk {
		panic(fmt.Errorf("Default builder allready exist: %s", vt.String()))
	}
}

func (c *factoryContext) getByNamePanic(ctx context.Context, name string, vt reflect.Type) any {
	if ret, err := c.getByName(ctx, name, vt); err != nil {
		panic(err)
	} else {
		return ret
	}
}

func (c *factoryContext) getByName(ctx context.Context, name string, vt reflect.Type) (any, error) {
	mb, ok := c.namedMap.Load(name)

	if ok {
		ctx = pushGetter(ctx, mb)
		defer popGetter(ctx)

		result := mb.getter(ctx)
		if vt != nil {
			rt := reflect.TypeOf(result)
			if vt.ConvertibleTo(rt) {
				return result, nil
			}
		} else {
			return result, nil
		}
	}

	return nil, fmt.Errorf("Named builder %s not found.", name)
}

func (c *factoryContext) setByName(name string, cci *contextCachedItem) {
	_, getOk := c.namedMap.LoadOrStore(name, cci)
	if getOk {
		panic(fmt.Errorf("Named builder allready exist: %s", name))
	}
}

func (c *factoryContext) evalExpr(ctx context.Context, code string) (any, error) {
	tree, _ := parser.Parse(code)

	exprCtx := &exprContext{ctx: ctx}
	ast.Walk(&tree.Node, exprCtx)

	c.exprEnvLock.RLock()
	defer c.exprEnvLock.RUnlock()

	return expr.Eval(code, c.exprEnvMap)
}

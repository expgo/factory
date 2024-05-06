package factory

import (
	"context"
	"fmt"
	"github.com/expgo/structure"
	"github.com/expgo/sync"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"reflect"
	"time"
)

var _context = &factoryContext{
	typedMap:     make(map[reflect.Type]*contextCachedItem),
	typedMapLock: sync.NewRWMutex(),
	namedMap:     make(map[string]*contextCachedItem),
	namedMapLock: sync.NewRWMutex(),
	exprEnvMap:   make(map[string]any),
	exprEnvLock:  sync.NewRWMutex(),
}

type factoryContext struct {
	typedMap     map[reflect.Type]*contextCachedItem // package:name -> must builder
	typedMapLock sync.RWMutex
	namedMap     map[string]*contextCachedItem // name -> must builder
	namedMapLock sync.RWMutex
	exprEnvMap   map[string]any
	exprEnvLock  sync.RWMutex
}

type contextCachedItem struct {
	_type  reflect.Type
	getter func(ctx context.Context) any
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
	return findTimeout(reflect.TypeOf((*T)(nil)), Opts.Timeout).(*T)
}

func FindTimeout[T any](timeout time.Duration) *T {
	return findTimeout(reflect.TypeOf((*T)(nil)), timeout).(*T)
}

func findTimeout(vt reflect.Type, timeout time.Duration) any {
	result := _context.getByType(getTimeoutContext(timeout), vt)

	resultType := reflect.TypeOf(result)
	if resultType.Kind() == reflect.Ptr && resultType.ConvertibleTo(vt) {
		return result
	}

	// panic
	panic(fmt.Errorf("Invalid type: need %v, get %v", vt, resultType))
}

func FindByName[T any](name string) *T {
	return findByNameTimeout(reflect.TypeOf((*T)(nil)), name, Opts.Timeout).(*T)
}

func FindByNameTimeout[T any](name string, timeout time.Duration) *T {
	return findByNameTimeout(reflect.TypeOf((*T)(nil)), name, timeout).(*T)
}

func findByNameTimeout(vt reflect.Type, name string, timeout time.Duration) any {
	result := _context.getByNamePanic(getTimeoutContext(timeout), name, vt)

	resultType := reflect.TypeOf(result)
	if resultType.Kind() == reflect.Ptr && resultType.ConvertibleTo(vt) {
		return result
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

	_context.typedMapLock.RLock()
	clonedMap := structure.CloneMap(_context.typedMap)
	_context.typedMapLock.RUnlock()

	for k, v := range clonedMap {
		if k.ConvertibleTo(vt) {
			rangeFunc(v.getter(ctx))
		}
	}
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
	c.typedMapLock.RLock()
	mb, ok := c.typedMap[vt]
	c.typedMapLock.RUnlock()

	if ok {
		ctx = pushGetter(ctx, mb)
		defer popGetter(ctx)

		return mb.getter(ctx)
	}

	if vt.Kind() == reflect.Interface {
		// 需求是接口才使用下面方法找寻
		c.typedMapLock.RLock()
		clonedMap := structure.CloneMap(c.typedMap)
		c.typedMapLock.RUnlock()

		convertibleMap := make(map[reflect.Type]*contextCachedItem)
		for k, v := range clonedMap {
			if k.ConvertibleTo(vt) {
				convertibleMap[k] = v
			}
		}

		convertibleMapSize := len(convertibleMap)

		if convertibleMapSize > 1 {
			panic(fmt.Errorf("Multiple default builders found for type: %v, please use named singleton", vt))
		}

		if convertibleMapSize == 1 {
			for _, v := range convertibleMap {
				mb = v
				ok = true
				break
			}

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
	c.typedMapLock.Lock()
	defer c.typedMapLock.Unlock()

	_, getOk := c.typedMap[vt]
	if getOk {
		panic(fmt.Errorf("Default builder allready exist: %s", vt.String()))
	}

	c.typedMap[vt] = cci
}

func (c *factoryContext) getByNamePanic(ctx context.Context, name string, vt reflect.Type) any {
	if ret, err := c.getByName(ctx, name, vt); err != nil {
		panic(err)
	} else {
		return ret
	}
}

func (c *factoryContext) getByName(ctx context.Context, name string, vt reflect.Type) (any, error) {
	c.namedMapLock.RLock()
	mb, ok := c.namedMap[name]
	c.namedMapLock.RUnlock()

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
	c.namedMapLock.Lock()
	defer c.namedMapLock.Unlock()

	_, getOk := c.namedMap[name]
	if getOk {
		panic(fmt.Errorf("Named builder allready exist: %s", name))
	}

	c.namedMap[name] = cci
}

func (c *factoryContext) evalExpr(ctx context.Context, code string) (any, error) {
	tree, _ := parser.Parse(code)

	exprCtx := &exprContext{ctx: ctx}
	ast.Walk(&tree.Node, exprCtx)

	c.exprEnvLock.RLock()
	defer c.exprEnvLock.RUnlock()

	return expr.Eval(code, c.exprEnvMap)
}

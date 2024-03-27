package factory

import (
	"fmt"
	"github.com/expgo/generic"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"reflect"
	"sync"
)

var _context = &context{}

type Getter func() any

type context struct {
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
	lock sync.RWMutex
}

func (c *exprEnv) Visit(node *ast.Node) {
	if s, ok := (*node).(*ast.IdentifierNode); ok {
		_, ok = c.getValue(s.String())
		if !ok {
			value := _context.getByName(s.String())
			c.setValue(s.String(), value)
		}
	}
}

func (c *exprEnv) getValue(name string) (any, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value, ok := c.env[name]
	return value, ok
}

func (c *exprEnv) setValue(name string, value any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.env[name] = value
}

func Find[T any]() *T {
	vt := reflect.TypeOf((*T)(nil))

	result := _context.getByType(vt)

	resultType := reflect.TypeOf(result)
	if resultType.Kind() == reflect.Ptr && resultType.ConvertibleTo(vt) {
		return result.(*T)
	}

	// panic
	panic(fmt.Sprintf("Invalid type: need %v, get %v", vt, resultType))
}

func FindByName[T any](name string) *T {
	vt := reflect.TypeOf((*T)(nil))

	result := _context.getByName(name)

	resultType := reflect.TypeOf(result)
	if resultType.Kind() == reflect.Ptr && resultType.ConvertibleTo(vt) {
		return result.(*T)
	}

	// panic
	panic(fmt.Sprintf("Invalid type: need %v, get %v", vt, resultType))
}

func Range[T any](rangeFunc func(any) bool) {
	vt := reflect.TypeOf((*T)(nil))

	if vt.Elem().Kind() == reflect.Interface {
		vt = vt.Elem()
	} else if vt.Elem().Kind() != reflect.Struct {
		panic("Range only range type and interface")
	}

	_context.defaultMustBuilderCache.Range(func(k reflect.Type, v *contextCachedItem) bool {
		if k.ConvertibleTo(vt) {
			return rangeFunc(v.getter())
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

func (c *context) getByNameOrType(name string, vt reflect.Type) any {
	mb, ok := c.namedMustBuilderCache.Load(name)

	if ok {
		result := mb.getter()
		rt := reflect.TypeOf(result)
		if vt.ConvertibleTo(rt) {
			return result
		}
	}

	return c.getByType(vt)
}

func (c *context) getByType(vt reflect.Type) any {
	mb, ok := c.defaultMustBuilderCache.Load(vt)

	if ok {
		return mb.getter()
	}

	if vt.Kind() == reflect.Interface {
		// 需求是接口才使用下面方法找寻
		convertibleList := c.defaultMustBuilderCache.Filter(func(k reflect.Type, v *contextCachedItem) bool {
			return k.ConvertibleTo(vt)
		})

		convertibleListSize := convertibleList.Size()

		if convertibleListSize > 1 {
			panic(fmt.Sprintf("Multiple default builders found for type: %v, please use named singleton", vt))
		}

		if convertibleListSize == 1 {
			convertibleList.Range(func(k reflect.Type, v *contextCachedItem) bool {
				mb = v
				ok = true
				return false
			})

			if ok {
				return mb.getter()
			}
		}
	}

	svt := vt
	if svt.Kind() == reflect.Ptr {
		svt = svt.Elem()
	}

	panic(fmt.Sprintf("use type to get Getter, %s:%s not found.", svt.PkgPath(), svt.Name()))

}

func (c *context) setByType(vt reflect.Type, builder Getter) {
	_, getOk := c.defaultMustBuilderCache.LoadOrStore(vt, &contextCachedItem{_type: vt, getter: builder})
	if getOk {
		panic(fmt.Sprintf("Default builder allready exist: %s", vt.String()))
	}
}

func (c *context) getByName(name string) any {
	mb, ok := c.namedMustBuilderCache.Load(name)

	if ok {
		return mb.getter()
	}

	panic(fmt.Sprintf("Named builder %s not found.", name))
}

func (c *context) setByName(name string, vt reflect.Type, builder Getter) {
	_, getOk := c.namedMustBuilderCache.LoadOrStore(name, &contextCachedItem{_type: vt, getter: builder})
	if getOk {
		panic(fmt.Sprintf("Named builder allready exist: %s", name))
	}
}

func (c *context) wiring(vt reflect.Type) {
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}

	_, ok := c.wiringCache.LoadOrStore(vt, true)
	if ok {
		panic(fmt.Sprintf("%s:%s is wiring, possible circular reference exists.", vt.PkgPath(), vt.Name()))
	}
}

func (c *context) wired(vt reflect.Type) {
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}

	c.wiringCache.Delete(vt)
}

func (c *context) evalExpr(code string) (any, error) {
	c.exprEnvInitOnce.Do(func() {
		c.exprEnv = &exprEnv{
			env: make(map[string]any),
		}
	})

	tree, _ := parser.Parse(code)
	ast.Walk(&tree.Node, c.exprEnv)

	c.exprEnv.lock.RLock()
	defer c.exprEnv.lock.RUnlock()

	return expr.Eval(code, c.exprEnv.env)
}

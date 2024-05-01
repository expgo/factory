package factory

import (
	"context"
	"fmt"
	"github.com/expgo/structure"
	"github.com/expgo/sync"
	"reflect"
)

const NewMethodName = "New"

var factories = make(map[reflect.Type]*_factory)
var factoriesLock = sync.NewRWMutex()

type _factory struct {
	factory     any
	factoryType reflect.Type
	returnType  reflect.Type
	methodName  string
	params      []string
}

func (f *_factory) MethodName(methodName string) *_factory {
	if f.factoryType.Kind() == reflect.Func {
		panic("factory is func, couldn't call MethodName")
	}

	if len(methodName) > 0 {
		f.methodName = methodName
	}

	return f
}

func (f *_factory) Params(params ...string) *_factory {
	f.params = params
	return f
}

func (f *_factory) CheckValid() {
	ft := f.factoryType
	vt := f.returnType
	if ft.Kind() == reflect.Func {
		if ft.NumOut() != 1 {
			panic(fmt.Errorf("func factory must return one value: %s", vt.String()))
		}

		if !ft.Out(0).AssignableTo(vt) {
			panic("func factory's return can't assign to the register type")
		}

		if ft.NumIn() != len(f.params) {
			panic("func factory params is not equals with params in")
		}
	} else if ft.Kind() == reflect.Ptr && ft.Elem().Kind() == reflect.Struct {
		newMethod, ok := ft.MethodByName(f.methodName)
		if !ok {
			panic(fmt.Errorf("can't find new method from factory of type %s", vt.String()))
		}

		if newMethod.Type.NumOut() != 1 {
			panic(fmt.Errorf("*struct factory must return one value: %s", vt.String()))
		}

		if !newMethod.Type.Out(0).AssignableTo(vt) {
			panic("*struct factory's return can't assign to the register type")
		}

		if newMethod.Type.NumIn() != len(f.params)+1 {
			panic("*struct factory params is not equals with params in")
		}
	} else {
		panic(fmt.Errorf("factory %s input must be a *struct or a func", vt.String()))
	}
}

func Factory[T any](f any) *_factory {
	vt := reflect.TypeOf((*T)(nil))
	if vt.Elem().Kind() == reflect.Interface {
		vt = vt.Elem()
	}

	fac := &_factory{
		factory:     f,
		factoryType: reflect.TypeOf(f),
		returnType:  vt,
		methodName:  NewMethodName,
	}

	factoriesLock.RLock()
	_, loaded := factories[vt]
	factoriesLock.RUnlock()

	if loaded {
		panic(fmt.Errorf("factory already exist: %s", vt.String()))
	}

	factoriesLock.Lock()
	factories[vt] = fac
	factoriesLock.Unlock()

	return fac
}

func callFactory(ctx context.Context, f *_factory, self any, fieldValue reflect.Value, structField reflect.StructField, newParams []string) error {
	vt := f.factoryType

	if newParams == nil || len(newParams) != len(f.params) {
		newParams = f.params
	}

	if vt.Kind() == reflect.Func {
		funcValue := reflect.ValueOf(f.factory)
		funcType := funcValue.Type()

		params, err := _getMethodParams(ctx, self, funcType, newParams, funcType.Name())
		if err != nil {
			panic(fmt.Errorf("factory func %s error: %v", structField.Type.String(), err))
		}

		values := funcValue.Call(params)

		return structure.SetField(fieldValue, values[0].Interface())
	} else {
		newMethod, ok := vt.MethodByName(f.methodName)
		if ok {
			if newMethod.Type.NumOut() != 1 {
				panic("new method must only return one value")
			}

			params, err := _getMethodParams(ctx, self, newMethod.Type, newParams, newMethod.Name)
			if err != nil {
				panic(fmt.Errorf("factory new %s error: %v", structField.Type.String(), err))
			}

			values := newMethod.Func.Call(append([]reflect.Value{reflect.ValueOf(f.factory)}, params...))

			return structure.SetField(fieldValue, values[0].Interface())
		}

		return fmt.Errorf("can't find new method from factory of type %s", structField.Type.String())
	}
}

package factory

import (
	"fmt"
	"github.com/expgo/generic"
	"github.com/expgo/structure"
	"reflect"
)

const NewMethodName = "New"

var factories = generic.Map[reflect.Type, any]{}

func RegisterFactory[T any](f any) {
	vt := reflect.TypeOf((*T)(nil))
	if vt.Elem().Kind() == reflect.Interface {
		vt = vt.Elem()
	}

	ft := reflect.TypeOf(f)
	if (ft.Kind() == reflect.Ptr && ft.Elem().Kind() == reflect.Struct) || ft.Kind() == reflect.Func {
		_, loaded := factories.LoadOrStore(vt, f)
		if loaded {
			panic(fmt.Errorf("factory already exist: %s", vt.String()))
		}
	} else {
		panic(fmt.Errorf("factory %s input must be a *struct or a func", vt.String()))
	}
}

func callFactory(f any, self any, fieldValue reflect.Value, structField reflect.StructField, newParams []string) error {
	vt := reflect.TypeOf(f)

	if vt.Kind() == reflect.Func {
		funcValue := reflect.ValueOf(f)
		funcType := funcValue.Type()

		params, err := _getMethodParams(self, funcType, newParams, funcType.Name())
		if err != nil {
			panic(fmt.Sprintf("factory func %s error: %v", structField.Type.String(), err))
		}

		values := funcValue.Call(params)

		return structure.SetField(fieldValue, values[0].Interface())
	} else {
		newMethod, ok := vt.MethodByName(NewMethodName)
		if ok {
			if newMethod.Type.NumOut() != 1 {
				panic("new method must only return one value")
			}

			params, err := _getMethodParams(self, newMethod.Type, newParams, newMethod.Name)
			if err != nil {
				panic(fmt.Sprintf("factory new %s error: %v", structField.Type.String(), err))
			}

			values := newMethod.Func.Call(append([]reflect.Value{reflect.ValueOf(f)}, params...))

			return structure.SetField(fieldValue, values[0].Interface())
		}

		return fmt.Errorf("can't find new method from factory of type %s", structField.Type.String())
	}
}

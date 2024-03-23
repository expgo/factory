package factory

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// DefaultInitMethodName is a constant representing the default name of the initialization method.
// It is used in the NewWithOption function to determine the name of the method to invoke during initialization.
// If the 'initMethodName' field in the Option struct is empty, the DefaultInitMethodName is used.
// If the 'useConstructor' field in the Option struct is true, the DefaultInitMethodName is set to the name of the struct.
// The DefaultInitMethodName is used in reflection to find and invoke the initialization method.
const DefaultInitMethodName = "Init"

type Option struct {
	useConstructor bool
	initMethodName string
	initParams     []string
	lock           sync.Mutex
}

func NewOption() *Option {
	return &Option{}
}

func (o *Option) UseConstructor(useConstructor bool) *Option {
	o.lock.Lock()
	defer o.lock.Unlock()

	if useConstructor && len(o.initMethodName) > 0 {
		panic("initMethodName must be empty when UseConstructor is true")
	}

	o.useConstructor = useConstructor
	return o
}

func (o *Option) InitMethodName(initMethodName string) *Option {
	o.lock.Lock()
	defer o.lock.Unlock()

	if len(initMethodName) > 0 && o.useConstructor {
		panic("useConstructor must be false when initMethodName is set")
	}

	o.initMethodName = initMethodName
	return o
}

func (o *Option) InitParams(initParams ...string) *Option {
	o.lock.Lock()
	defer o.lock.Unlock()

	if len(o.initParams) > 0 {
		panic("params already set")
	}

	o.initParams = initParams

	return o
}

var newDefaultOption = NewOption()

func New[T any]() *T {
	return NewWithOption[T](newDefaultOption)
}

func _getInitParams(self any, initMethod reflect.Method, option *Option) ([]reflect.Value, error) {
	var params []reflect.Value

	if len(option.initParams) == 0 {
		for i := 1; i < initMethod.Type.NumIn(); i++ {
			paramType := initMethod.Type.In(i)
			if (paramType.Kind() == reflect.Ptr && paramType.Elem().Kind() == reflect.Struct) || paramType.Kind() == reflect.Interface {
				params = append(params, reflect.ValueOf(_context.getByType(paramType)))
			} else {
				return nil, errors.New(fmt.Sprintf("Method %s's %d argument must be a struct point or an interface", initMethod.Name, i))
			}
		}
	} else if len(option.initParams)+1 == initMethod.Type.NumIn() {
		for i := 0; i < initMethod.Type.NumIn()-1; i++ {
			paramType := initMethod.Type.In(i + 1)

			tagValue, err := ParseTagValue(option.initParams[i], nil)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Method %s's %d argument tag is err: %v", initMethod.Name, i+1, err))
			}

			v, err := getValueByWireTag(self, tagValue, paramType)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Method %s's %d argument get value from tag err: %v", initMethod.Name, i+1, err))
			}

			params = append(params, reflect.ValueOf(v))
		}
	} else {
		return nil, errors.New("init params count must equals with method params count")
	}

	return params, nil
}

func NewWithOption[T any](option *Option) *T {
	if option == nil {
		option = newDefaultOption
	}

	t := new(T)

	vt := reflect.TypeOf(t)

	if vt.Kind() == reflect.Ptr && vt.Elem().Kind() == reflect.Struct {
		vte := vt.Elem()

		// get init method name
		initMethodName := option.initMethodName
		if len(initMethodName) == 0 {
			initMethodName = DefaultInitMethodName
		}
		if option.useConstructor {
			initMethodName = vte.Name()
		}

		// 确保方法的第一个字母为大写
		initMethodName = strings.ToTitle(initMethodName[:1]) + initMethodName[1:]

		// from name get method
		initMethod, ok := vt.MethodByName(initMethodName)
		if ok {
			if initMethod.Type.NumOut() > 0 {
				panic(fmt.Sprintf("Init method '%s' must not have return values", initMethodName))
			}

			params, err := _getInitParams(t, initMethod, option)
			if err == nil {
				defer initMethod.Func.Call(append([]reflect.Value{reflect.ValueOf(t)}, params...))
			} else {
				panic(fmt.Sprintf("Create %s error: %v", vte.Name(), err))
			}
		}
	} else {
		panic("T must be a struct type")
	}

	// do auto wire
	if err := AutoWire(t); err != nil {
		panic(fmt.Sprintf("Create %s error: %v", vt.Elem().Name(), err))
	}

	return t
}

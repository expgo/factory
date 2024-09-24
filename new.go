package factory

import (
	"context"
	"errors"
	"fmt"
	"github.com/expgo/sync"
	"reflect"
	"strings"
	"time"
)

// DefaultInitMethodName is a constant representing the default name of the initialization method.
// It is used in the NewWithOption function to determine the name of the method to invoke during initialization.
// If the 'initMethodName' field in the Option struct is empty, the DefaultInitMethodName is used.
// If the 'useConstructor' field in the Option struct is true, the DefaultInitMethodName is set to the name of the struct.
// The DefaultInitMethodName is used in reflection to find and invoke the initialization method.
const DefaultInitMethodName = "Init"

var newCtxMap = make(map[int]context.Context)
var newCtxMapLock = sync.NewRWMutex()

type Option struct {
	useConstructor bool
	initMethodName string
	initParams     []string
	lock           sync.Mutex
}

func NewOption() *Option {
	return &Option{
		lock: sync.NewMutex(),
	}
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
	return initWithOptionTimeout(new(T), newDefaultOption, Opts.Timeout, nil).(*T)
}

func NewWithOption[T any](option *Option) *T {
	return initWithOptionTimeout(new(T), option, Opts.Timeout, nil).(*T)
}

func NewWithFunc[T any](createFunc func() *T, beforeInit func(*T)) *T {
	return NewWithFuncOption[T](createFunc, beforeInit, newDefaultOption)
}

func NewWithFuncOption[T any](createFunc func() *T, beforeInit func(*T), option *Option) *T {
	var t *T
	if createFunc != nil {
		t = createFunc()
	} else {
		t = new(T)
	}

	return initWithOptionTimeout(t, option, Opts.Timeout, func() {
		if beforeInit != nil {
			beforeInit(t)
		}
	}).(*T)
}

func NewBeforeInit[T any](beforeInit func(*T)) *T {
	return NewBeforeInitOption(beforeInit, newDefaultOption)
}

func NewBeforeInitOption[T any](beforeInit func(*T), option *Option) *T {
	t := new(T)
	return initWithOptionTimeout(t, option, Opts.Timeout, func() {
		if beforeInit != nil {
			beforeInit(t)
		}
	}).(*T)
}

func initWithOptionTimeout(t any, option *Option, timeout time.Duration, beforeInit func()) any {
	goId := sync.GoId()
	newCtxMapLock.RLock()

	ctx, loaded := newCtxMap[goId]
	newCtxMapLock.RUnlock()

	if !loaded {
		ctx = initTypeCtx(getTimeoutContext(timeout))

		newCtxMapLock.Lock()
		newCtxMap[goId] = ctx
		newCtxMapLock.Unlock()

		defer func() {
			newCtxMapLock.Lock()
			delete(newCtxMap, goId)
			newCtxMapLock.Unlock()
		}()
	}

	return initWithOptionContext(t, ctx, option, beforeInit)
}

func initWithOptionContext(t any, ctx context.Context, option *Option, beforeInit func()) any {
	if option == nil {
		option = newDefaultOption
	}

	vt := reflect.TypeOf(t)
	ctx = pushType(ctx, vt)
	defer popType(ctx)

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
				panic(fmt.Errorf("init method '%s' must not have return values", initMethodName))
			}

			params, err := _getMethodParams(ctx, t, initMethod.Type, option.initParams, initMethod.Name)
			if err == nil {
				// defer 将init的调用放到auto wire之后
				defer initMethod.Func.Call(append([]reflect.Value{reflect.ValueOf(t)}, params...))
			} else {
				panic(fmt.Errorf("create %s error: %v", vte.Name(), err))
			}
		}
	} else {
		panic(errors.New("T must be a struct type"))
	}

	// do auto wire
	if err := autoWireContext(ctx, t); err != nil {
		panic(fmt.Errorf("create %s error: %v", vt.Elem().Name(), err))
	}

	if beforeInit != nil {
		beforeInit()
	}

	return t
}

func _getMethodParams(ctx context.Context, self any, methodType reflect.Type, methodParams []string, methodName string) ([]reflect.Value, error) {
	var params []reflect.Value

	baseIndex := methodType.NumIn() - len(methodParams)

	if len(methodParams) == 0 {
		for i := 1; i < methodType.NumIn(); i++ {
			paramType := methodType.In(i)
			if (paramType.Kind() == reflect.Ptr && paramType.Elem().Kind() == reflect.Struct) || paramType.Kind() == reflect.Interface {
				params = append(params, reflect.ValueOf(_context.getByType(ctx, paramType)))
			} else {
				return nil, fmt.Errorf("method %s's %d argument must be a struct point or an interface", methodName, i)
			}
		}
	} else if baseIndex == 0 || baseIndex == 1 {
		for i := 0; i < methodType.NumIn()-baseIndex; i++ {
			paramType := methodType.In(i + baseIndex)

			tagValue, err := ParseTagValue(methodParams[i], nil)
			if err != nil {
				return nil, fmt.Errorf("method %s's %d argument tag is err: %v", methodName, i+baseIndex, err)
			}

			v, err := getValueByWireTag(ctx, self, tagValue, paramType)
			if err != nil {
				return nil, fmt.Errorf("method %s's %d argument get value from tag err: %v", methodName, i+baseIndex, err)
			}

			params = append(params, reflect.ValueOf(v))
		}
	} else {
		return nil, errors.New("init params count must equals with method params count")
	}

	return params, nil
}

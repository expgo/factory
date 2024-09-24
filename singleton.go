package factory

import (
	"context"
	"fmt"
	"github.com/expgo/sync"
	"reflect"
	"strings"
	"time"
)

type singleton struct {
	once     sync.Once
	obj      any
	initFunc func() any
	option   Option

	_name string
	lock  sync.Mutex

	cci *contextCachedItem
}

func _singletonWithType(vt reflect.Type) *singleton {
	result := &singleton{
		once: sync.NewOnce(),
		lock: sync.NewMutex(),
		option: Option{
			lock: sync.NewMutex(),
		},
	}

	result.cci = &contextCachedItem{}

	result.cci._type = vt

	result.obj = reflect.New(vt.Elem()).Interface()

	result.cci.getter = func(ctx context.Context) any {
		return result.getWithContext(ctx)
	}

	return result
}

func Singleton[T any]() *singleton {
	return _singletonWithType(reflect.TypeOf((*T)(nil))).setType()
}

func NamedSingleton[T any](name string) *singleton {
	return _singletonWithType(reflect.TypeOf((*T)(nil))).Name(name)
}

func (s *singleton) setType() *singleton {
	_context.setByType(s.cci._type, s.cci)
	return s
}

func (s *singleton) Name(name string) *singleton {
	s.lock.Lock()
	defer s.lock.Unlock()

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		panic("name must not be empty")
	}

	if len(s._name) == 0 {
		_context.setByName(name, s.cci)
		s._name = name
	} else {
		panic("name already set")
	}

	return s
}

func (s *singleton) WithOption(option *Option) *singleton {
	s.lock.Lock()
	defer s.lock.Unlock()

	if option != nil {
		s.option.lock.Lock()
		defer s.option.lock.Unlock()

		s.option.useConstructor = option.useConstructor
		s.option.initMethodName = option.initMethodName
		s.option.initParams = option.initParams
	}

	return s
}

func (s *singleton) SetInitFunc(initFunc func() any) *singleton {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.initFunc = initFunc
	return s
}

func (s *singleton) UseConstructor(useConstructor bool) *singleton {
	s.option.UseConstructor(useConstructor)
	return s
}

func (s *singleton) InitMethodName(initMethodName string) *singleton {
	s.option.InitMethodName(initMethodName)
	return s
}

func (s *singleton) InitParams(initParams ...string) *singleton {
	s.option.InitParams(initParams...)
	return s
}

func (s *singleton) getWithContext(ctx context.Context) any {
	timeout := getContextTimeout(ctx)
	err := s.once.DoTimeout(timeout, func() error {
		if s.initFunc != nil {
			s.obj = s.initFunc()
		} else {
			s.obj = initWithOptionContext(s.obj, getNextTimeoutContext(ctx), &s.option, nil)
		}

		return nil
	})

	if err != nil {
		panic(fmt.Errorf("[%s]init singleton %s, timeout: %s err: %+v", time.Now(), s.cci._type.String(), timeout, err))
	}

	return s.obj
}

func Getter[T any](s *singleton) func() *T {
	return func() *T {
		return s.getWithContext(getTimeoutContext(Opts.Timeout)).(*T)
	}
}

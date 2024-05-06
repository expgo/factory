package factory

import (
	"context"
	"fmt"
	"github.com/expgo/sync"
	"reflect"
	"strings"
	"time"
)

type singleton[T any] struct {
	once     sync.Once
	obj      *T
	initFunc func() *T
	option   Option

	_name string
	lock  sync.Mutex

	cci *contextCachedItem
}

func _singleton[T any]() *singleton[T] {
	result := &singleton[T]{
		once: sync.NewOnce(),
		lock: sync.NewMutex(),
		option: Option{
			lock: sync.NewMutex(),
		},
	}

	result.cci = &contextCachedItem{}

	result.cci._type = reflect.TypeOf((*T)(nil))

	result.cci.getter = func(ctx context.Context) any {
		return result.getWithContext(ctx)
	}

	return result
}

func Singleton[T any]() *singleton[T] {
	result := _singleton[T]()

	_context.setByType(result.cci._type, result.cci)

	return result
}

func NamedSingleton[T any](name string) *singleton[T] {
	result := _singleton[T]()

	return result.Name(name)
}

func (s *singleton[T]) Name(name string) *singleton[T] {
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

func (s *singleton[T]) WithOption(option *Option) *singleton[T] {
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

func (s *singleton[T]) SetInitFunc(initFunc func() *T) *singleton[T] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.initFunc = initFunc
	return s
}

func (s *singleton[T]) UseConstructor(useConstructor bool) *singleton[T] {
	s.option.UseConstructor(useConstructor)
	return s
}

func (s *singleton[T]) InitMethodName(initMethodName string) *singleton[T] {
	s.option.InitMethodName(initMethodName)
	return s
}

func (s *singleton[T]) InitParams(initParams ...string) *singleton[T] {
	s.option.InitParams(initParams...)
	return s
}

func (s *singleton[T]) Get() *T {
	return s.getWithContext(getTimeoutContext(Opts.Timeout))
}

func (s *singleton[T]) getWithContext(ctx context.Context) *T {
	timeout := getContextTimeout(ctx)
	err := s.once.DoTimeout(timeout, func() error {
		if s.initFunc != nil {
			s.obj = s.initFunc()
		} else {
			s.obj = initWithOptionContext(new(T), getNextTimeoutContext(ctx), &s.option).(*T)
		}

		return nil
	})

	if err != nil {
		panic(fmt.Errorf("[%s]init singleton %s, timeout: %s err: %+v", time.Now(), s.cci._type.String(), timeout, err))
	}

	return s.obj
}

func (s *singleton[T]) Getter() func() *T {
	return func() *T {
		return s.Get()
	}
}

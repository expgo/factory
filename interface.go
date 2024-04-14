package factory

import (
	"context"
	"fmt"
	"github.com/expgo/sync"
	"reflect"
	"strings"
)

type iInterface[T any] struct {
	once     sync.Once
	obj      T
	initFunc func() T

	_name   string
	_type   reflect.Type
	_getter Getter
	lock    sync.Mutex
}

func _interface[T any]() *iInterface[T] {
	result := &iInterface[T]{
		once: sync.NewOnce(),
		lock: sync.NewMutex(),
	}

	result._type = reflect.TypeOf((*T)(nil)).Elem()

	result._getter = func(ctx context.Context) any {
		return result.getWithContext(ctx)
	}

	return result
}

func Interface[T any]() *iInterface[T] {
	result := _interface[T]()

	_context.setByType(result._type, result._getter)

	return result
}

func NamedInterface[T any](name string) *iInterface[T] {
	result := _interface[T]()

	return result.Name(name)
}

func (s *iInterface[T]) Name(name string) *iInterface[T] {
	s.lock.Lock()
	defer s.lock.Unlock()

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		panic("name must not be empty")
	}

	if len(s._name) == 0 {
		_context.setByName(name, s._type, s._getter)
		s._name = name
	} else {
		panic("name already set")
	}

	return s
}

func (s *iInterface[T]) SetInitFunc(initFunc func() T) *iInterface[T] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.initFunc = initFunc
	return s
}

func (s *iInterface[T]) Get() T {
	return s.getWithContext(getTimeoutContext(Opts.DefaultTimeout))
}

func (s *iInterface[T]) getWithContext(ctx context.Context) T {
	timeout := getContextTimeout(ctx)
	err := s.once.DoTimeout(timeout, func() error {
		if s.initFunc != nil {
			s.obj = s.initFunc()
		} else {
			panic("initFunc must be set")
		}
		return nil
	})

	if err != nil {
		panic(fmt.Errorf("init interface %s, timeout: %s err: %+v", s._type.String(), timeout, err))
	}

	return s.obj
}

func (s *iInterface[T]) Getter() func() T {
	return func() T {
		return s.Get()
	}
}

package factory

import (
	"reflect"
	"strings"
	"sync"
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
	result := &iInterface[T]{}

	result._type = reflect.TypeOf((*T)(nil)).Elem()

	result._getter = func() any {
		return result.Get()
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
	s.once.Do(func() {
		if s.initFunc != nil {
			s.obj = s.initFunc()
		} else {
			panic("initFunc must be set")
		}
	})

	return s.obj
}

func (s *iInterface[T]) Getter() func() T {
	return func() T {
		return s.Get()
	}
}

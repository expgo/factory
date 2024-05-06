package factory

import (
	"context"
	"fmt"
	"github.com/expgo/sync"
	"reflect"
	"strings"
)

type iInterface struct {
	once     sync.Once
	obj      any
	initFunc func() any

	_name string
	lock  sync.Mutex

	cci *contextCachedItem
}

func _interfaceWithType(vt reflect.Type) *iInterface {
	result := &iInterface{
		once: sync.NewOnce(),
		lock: sync.NewMutex(),
	}

	result.cci = &contextCachedItem{}

	result.cci._type = vt

	result.cci.getter = func(ctx context.Context) any {
		return result.getWithContext(ctx)
	}

	return result
}

func Interface[T any]() *iInterface {
	return _interfaceWithType(reflect.TypeOf((*T)(nil)).Elem()).setType()
}

func NamedInterface[T any](name string) *iInterface {
	return _interfaceWithType(reflect.TypeOf((*T)(nil)).Elem()).Name(name)
}

func (s *iInterface) setType() *iInterface {
	_context.setByType(s.cci._type, s.cci)
	return s
}

func (s *iInterface) Name(name string) *iInterface {
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

func (s *iInterface) SetInitFunc(initFunc func() any) *iInterface {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.initFunc = initFunc
	return s
}

func (s *iInterface) getWithContext(ctx context.Context) any {
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
		panic(fmt.Errorf("init interface %s, timeout: %s err: %+v", s.cci._type.String(), timeout, err))
	}

	return s.obj
}

func IGetter[T any](s *iInterface) func() T {
	return func() T {
		return s.getWithContext(getTimeoutContext(Opts.Timeout)).(T)
	}
}

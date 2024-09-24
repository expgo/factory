package factory

import (
	"reflect"
	"sync"
)

var _poolCache = map[reflect.Type]*sync.Pool{}
var _poolCacheLock = &sync.RWMutex{}

func getPoolByType(vt reflect.Type) *sync.Pool {
	_poolCacheLock.RLock()
	pool, ok := _poolCache[vt]
	_poolCacheLock.RUnlock()

	if !ok {
		pool = &sync.Pool{
			New: func() interface{} {
				return initWithOptionTimeout(reflect.New(vt).Interface(), newDefaultOption, Opts.Timeout, nil)
			},
		}

		_poolCacheLock.Lock()
		_poolCache[vt] = pool
		_poolCacheLock.Unlock()
	}

	return pool
}

func Get[T any]() *T {
	return getPoolByType(reflect.TypeOf((*T)(nil)).Elem()).Get().(*T)
}

func Put[T any](t *T) {
	if t == nil {
		return
	}

	getPoolByType(reflect.TypeOf((*T)(nil)).Elem()).Put(t)
}

func SetPoolInit[T any](option *Option) {
	setPoolInitWithType(reflect.TypeOf((*T)(nil)).Elem(), option)
}

func setPoolInitWithType(vt reflect.Type, option *Option) {
	pool := &sync.Pool{
		New: func() interface{} {
			return initWithOptionTimeout(reflect.New(vt).Interface(), option, Opts.Timeout, nil)
		},
	}

	_poolCacheLock.Lock()
	defer _poolCacheLock.Unlock()

	_poolCache[vt] = pool
}

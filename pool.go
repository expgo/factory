package factory

import (
	"github.com/expgo/generic"
	"reflect"
	"sync"
)

var _poolCache = generic.Cache[reflect.Type, *sync.Pool]{}
var _poolCacheLock = sync.RWMutex{}

func Get[T any]() *T {
	_poolCacheLock.RLock()
	defer _poolCacheLock.RUnlock()

	poolType := reflect.TypeOf((*T)(nil)).Elem()
	pool, _ := _poolCache.GetOrLoad(poolType, func(_ reflect.Type) (*sync.Pool, error) {
		return &sync.Pool{
			New: func() interface{} {
				return New[T]()
			},
		}, nil
	})
	return pool.Get().(*T)
}

func Put[T any](t *T) {
	if t == nil {
		return
	}

	_poolCacheLock.RLock()
	defer _poolCacheLock.RUnlock()

	poolType := reflect.TypeOf((*T)(nil)).Elem()
	pool, _ := _poolCache.GetOrLoad(poolType, func(_ reflect.Type) (*sync.Pool, error) {
		return &sync.Pool{
			New: func() interface{} {
				return New[T]()
			},
		}, nil
	})
	pool.Put(t)
}

//func SetPoolInit[T any](option *Option) {
//	_poolCacheLock.Lock()
//	defer _poolCacheLock.Unlock()
//
//	poolType := reflect.TypeOf((*T)(nil)).Elem()
//	_poolCache.Evict(poolType)
//	_poolCache.Set(poolType, &sync.Pool{
//		New: func() interface{} {
//			return NewWithOption[T](option)
//		},
//	})
//}

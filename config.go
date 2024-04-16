package factory

import (
	"context"
	"fmt"
	"github.com/expgo/generic"
	"os"
	"reflect"
	"strconv"
	"time"
)

const TimeoutKey = "Timeout"
const GetterKey = "Getter"
const TypeKey = "type"

var Opts = struct {
	EnableTimeout   bool
	Timeout         time.Duration
	TimeoutInterval time.Duration
	Log             Logger
}{
	EnableTimeout:   false,
	Timeout:         3 * time.Second,
	TimeoutInterval: 100 * time.Millisecond,
	Log:             &logger{},
}

func init() {
	if b, err := strconv.ParseBool(os.Getenv("FACTORY_ENABLE_TIMEOUT")); err == nil {
		Opts.EnableTimeout = b
		Opts.Log.Debugf("EnableTimeout set to %v", b)
	}

	if n, _ := strconv.Atoi(os.Getenv("FACTORY_TIMEOUT")); n > 0 {
		Opts.Timeout = time.Duration(n) * time.Second
		Opts.Log.Debugf("Timeout set to %v", Opts.Timeout)
	}

	if n, _ := strconv.Atoi(os.Getenv("FACTORY_TIMEOUT_INTERVAL")); n > 0 {
		Opts.TimeoutInterval = time.Duration(n) * time.Millisecond
		Opts.Log.Debugf("TimeoutInterval set to %v", Opts.TimeoutInterval)
	}
}

func getTimeoutContext(timeout time.Duration) context.Context {
	return context.WithValue(context.Background(), TimeoutKey, timeout)
}

func getNextTimeoutContext(ctx context.Context) context.Context {
	if Opts.EnableTimeout {
		duration := getContextTimeout(ctx)
		duration -= Opts.TimeoutInterval
		if duration <= 0 {
			panic("need larger DefaultFindTimeout")
		}
		return context.WithValue(ctx, TimeoutKey, duration)
	} else {
		return context.WithValue(ctx, TimeoutKey, 0)
	}
}

func getContextTimeout(ctx context.Context) time.Duration {
	if Opts.EnableTimeout {
		value := ctx.Value(TimeoutKey)
		if value != nil {
			return value.(time.Duration)
		} else {
			return Opts.Timeout
		}
	} else {
		return 0
	}
}

func pushGetter(ctx context.Context, ci *contextCachedItem) context.Context {
	ciSetValue := ctx.Value(GetterKey)
	if ciSetValue == nil {
		ciSetValue = &generic.Set[*contextCachedItem]{}
		ctx = context.WithValue(ctx, GetterKey, ciSetValue)
	}

	ciSet := ciSetValue.(*generic.Set[*contextCachedItem])
	if !ciSet.Add(ci) {
		panic(fmt.Errorf("getting %s, possible circular reference with %s", ci._type.String(), lastGetter(ctx)))
	}
	return ctx
}

func popGetter(ctx context.Context) {
	ciSetValue := ctx.Value(GetterKey)
	if ciSetValue != nil {
		ciSet := ciSetValue.(*generic.Set[*contextCachedItem])
		idx := ciSet.Size() - 1
		if idx >= 0 {
			ciSet.RemoveAt(idx)
		}
	}
}

func lastGetter(ctx context.Context) string {
	ciSetValue := ctx.Value(GetterKey)
	if ciSetValue != nil {
		ciSet := ciSetValue.(*generic.Set[*contextCachedItem])
		lastIdx := ciSet.Size() - 1
		if lastIdx >= 0 {
			ci, err := ciSet.At(lastIdx)
			if err != nil {
				return ""
			}
			return ci._type.String()
		}
	}
	return ""
}

func initTypeCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, TypeKey, &generic.Set[reflect.Type]{})
}

func pushType(ctx context.Context, _type reflect.Type) context.Context {
	typeSetValue := ctx.Value(TypeKey)
	if typeSetValue == nil {
		typeSetValue = &generic.Set[reflect.Type]{}
		ctx = context.WithValue(ctx, TypeKey, typeSetValue)
	}

	typeSet := typeSetValue.(*generic.Set[reflect.Type])
	if !typeSet.Add(_type) {
		panic(fmt.Errorf("getting %s, possible circular reference with %s", _type.String(), lastType(ctx)))
	}
	return ctx
}

func popType(ctx context.Context) {
	typeSetValue := ctx.Value(TypeKey)
	if typeSetValue != nil {
		ciSet := typeSetValue.(*generic.Set[reflect.Type])
		idx := ciSet.Size() - 1
		if idx >= 0 {
			ciSet.RemoveAt(idx)
		}
	}
}

func lastType(ctx context.Context) string {
	typeSetValue := ctx.Value(TypeKey)
	if typeSetValue != nil {
		typeSet := typeSetValue.(*generic.Set[reflect.Type])
		lastIdx := typeSet.Size() - 1
		if lastIdx >= 0 {
			_type, err := typeSet.At(lastIdx)
			if err != nil {
				return ""
			}
			return _type.String()
		}
	}
	return ""
}

package factory

import (
	"context"
	"fmt"
	"github.com/expgo/generic"
	"time"
)

const TimeoutKey = "Timeout"
const GetterKey = "Getter"

var Opts = struct {
	EnableTimeout          bool
	DefaultTimeout         time.Duration
	DefaultTimeoutInterval time.Duration
}{
	EnableTimeout:          true,
	DefaultTimeout:         3 * time.Second,
	DefaultTimeoutInterval: 100 * time.Millisecond,
}

func getTimeoutContext(timeout time.Duration) context.Context {
	return context.WithValue(context.Background(), TimeoutKey, timeout)
}

func getNextTimeoutContext(ctx context.Context) context.Context {
	duration := getContextTimeout(ctx)
	duration -= Opts.DefaultTimeoutInterval
	if duration <= 0 {
		panic("need larger DefaultFindTimeout")
	}
	return context.WithValue(ctx, TimeoutKey, duration)
}

func getContextTimeout(ctx context.Context) time.Duration {
	if Opts.EnableTimeout {
		value := ctx.Value(TimeoutKey)
		if value != nil {
			return value.(time.Duration)
		} else {
			return Opts.DefaultTimeout
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

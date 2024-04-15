package factory

import (
	"context"
	"fmt"
	"github.com/expgo/generic"
	"reflect"
	"time"
)

const TimeoutKey = "Timeout"
const NameGettingKey = "NameGetting"
const TypeGettingKey = "TypeGetting"
const LastGettingKey = "LastGetting"

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

func pushNameGetting(ctx context.Context, name string) context.Context {
	nameSetValue := ctx.Value(NameGettingKey)
	if nameSetValue == nil {
		nameSetValue = &generic.Set[string]{}
		ctx = context.WithValue(ctx, NameGettingKey, nameSetValue)
	}

	nameSet := nameSetValue.(*generic.Set[string])
	if !nameSet.Add(name) {

		panic(fmt.Errorf("name %s is getting, possible circular reference with %s ", name, getLastGetting(ctx)))
	}

	return setLastGetting(ctx, fmt.Sprintf("Named: %s", name))
}

func popNameGetting(ctx context.Context, name string) {
	nameSetValue := ctx.Value(NameGettingKey)
	if nameSetValue != nil {
		nameSet := nameSetValue.(*generic.Set[string])
		nameSet.Remove(name)
	}
}

func pushTypeGetting(ctx context.Context, vt reflect.Type) context.Context {
	typeSetValue := ctx.Value(TypeGettingKey)
	if typeSetValue == nil {
		typeSetValue = &generic.Set[reflect.Type]{}
		ctx = context.WithValue(ctx, TypeGettingKey, typeSetValue)
	}

	typeSet := typeSetValue.(*generic.Set[reflect.Type])
	if !typeSet.Add(vt) {
		panic(fmt.Errorf("type %s is getting, possible circular reference with %s", vt, getLastGetting(ctx)))
	}

	return setLastGetting(ctx, fmt.Sprintf("Typed: %s", vt.String()))
}

func popTypeGetting(ctx context.Context, vt reflect.Type) {
	typeSetValue := ctx.Value(TypeGettingKey)
	if typeSetValue != nil {
		typeSet := typeSetValue.(*generic.Set[reflect.Type])
		typeSet.Remove(vt)
	}
}

func getLastGetting(ctx context.Context) string {
	lastGettingValue := ctx.Value(LastGettingKey)
	if lastGettingValue != nil {
		if lastGetting, ok := lastGettingValue.(string); ok {
			return lastGetting
		}
	}

	return ""
}

func setLastGetting(ctx context.Context, lastGetting string) context.Context {
	return context.WithValue(ctx, LastGettingKey, lastGetting)
}

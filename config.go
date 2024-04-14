package factory

import (
	"context"
	"time"
)

const TimeoutKey = "Timeout"

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

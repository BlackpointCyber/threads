package threads

import (
	"context"
	"time"
)

type timeAfter func(d time.Duration) <-chan time.Time

type ctxTimeMockKey struct{}

func ContextWithTimeMock(ctx context.Context, t timeAfter) context.Context {
	return context.WithValue(ctx, ctxTimeMockKey{}, t)
}

func getTimeAfter(ctx context.Context) timeAfter {
	t, _ := ctx.Value(ctxTimeMockKey{}).(timeAfter)
	if t != nil {
		return t
	}
	return time.After
}

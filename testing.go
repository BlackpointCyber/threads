package threads

import (
	"context"
	"time"
)

type TimeAfterMock func(triggerCh chan time.Time, waitCh chan time.Duration)

type timeAfter func(d time.Duration) <-chan time.Time

type ctxTimeMockKey struct{}

func ContextWithTimeAfterMock(ctx context.Context, timeAfterMock TimeAfterMock) context.Context {
	waitCh := make(chan time.Duration, 10)
	triggerCh := make(chan time.Time)

	go timeAfterMock(triggerCh, waitCh)

	return context.WithValue(ctx, ctxTimeMockKey{}, timeAfter(func(d time.Duration) <-chan time.Time {
		waitCh <- d
		return triggerCh
	}))
}

func getTimeAfter(ctx context.Context) timeAfter {
	t, _ := ctx.Value(ctxTimeMockKey{}).(timeAfter)
	if t != nil {
		return t
	}
	return time.After
}

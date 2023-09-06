package threads

import (
	"context"
	"fmt"
	"time"
)

type retryWorkerErr struct {
	error
	d time.Duration
}

func RetryWorkerIn(d time.Duration) error {
	return retryWorkerErr{
		error: fmt.Errorf("this worker will be executed again in %v", d),
		d:     d,
	}
}

type adjustIntervalErr struct {
	error
	d time.Duration
}

func AdjustInterval(d time.Duration) error {
	return adjustIntervalErr{
		error: fmt.Errorf("adjusting worker interval to: %v", d),
		d:     d,
	}
}

type intervalType interface {
	time.Duration | func() time.Duration
}

func intervalTypeToFunc(t any) func() time.Duration {
	switch dur := t.(type) {
	case time.Duration:
		return func() time.Duration {
			return dur
		}
	case func() time.Duration:
		return dur
	default:
		panic("code error: unsupported interval type received")
	}
}

func PeriodicWorker[T intervalType](
	baseIterationInterval T,
	doWork Worker,
) Worker {
	iterationInterval := intervalTypeToFunc(baseIterationInterval)

	return func(ctx context.Context) error {
		// This allows us to mock time.After using
		// the ContextWithTimeMock function:
		timeAfter := getTimeAfter(ctx)

		for {
			nextIteration := iterationInterval()

			err := doWork(ctx)
			if err != nil {
				switch knownErr := err.(type) {
				case retryWorkerErr:
					nextIteration = knownErr.d
				case adjustIntervalErr:
					iterationInterval = func() time.Duration {
						return knownErr.d
					}
				default:
					return err
				}
			}

			// Blocks until the next iteration
			// or until the context is cancelled:
			select {
			case <-ctx.Done():
				return nil
			case <-timeAfter(nextIteration):
			}
		}
	}
}

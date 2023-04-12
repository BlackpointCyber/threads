package threads

import (
	"context"
	"time"
)

func PeriodicWorker(
	iterationInterval time.Duration,
	doWork Worker,
) Worker {
	return func(ctx context.Context) error {
		// This allows us to mock time.After using
		// the ContextWithTimeMock function:
		timeAfter := getTimeAfter(ctx)

		for {
			err := doWork(ctx)
			if err != nil {
				return err
			}

			// Blocks until the next iteration
			// or until the context is cancelled:
			select {
			case <-ctx.Done():
				return nil
			case <-timeAfter(iterationInterval):
			}
		}
	}
}

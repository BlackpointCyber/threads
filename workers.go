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
			case <-time.After(iterationInterval):
			}
		}
	}
}

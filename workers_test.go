package threads

import (
	"context"
	"fmt"
	"testing"
	"time"

	tt "github.bpcyber.com/vgarcia/threads/internal/testtools"
)

func TestPeriodicWorker(t *testing.T) {
	ctx := context.Background()

	t.Run("should stop executing if an error is returned", func(t *testing.T) {
		var numCalls int
		doneCh := make(chan struct{}, 1)

		go PeriodicWorker(100*time.Microsecond, func(ctx context.Context) error {
			numCalls++
			close(doneCh)
			return fmt.Errorf("fake-error")
		})(ctx)

		<-doneCh

		// Give it time to do another iteration (to prove it really stopped):
		time.Sleep(1000 * time.Microsecond)

		tt.AssertEqual(t, numCalls, 1)
	})

	t.Run("should run the callback imediately right after starting", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Microsecond)
		defer cancel()

		var numCalls int
		doneCh := make(chan struct{}, 1)
		go PeriodicWorker(2000*time.Microsecond, func(ctx context.Context) error {
			numCalls++
			doneCh <- struct{}{}
			return nil
		})(ctx)

		<-doneCh

		// Give it time to do another iteration (to prove it really stopped):
		time.Sleep(2500 * time.Microsecond)

		tt.AssertEqual(t, numCalls, 1)
	})

	t.Run("should run the callback periodically", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 5200*time.Microsecond)
		defer cancel()

		var numCalls int
		go PeriodicWorker(5000*time.Microsecond, func(ctx context.Context) error {
			numCalls++
			return nil
		})(ctx)

		// Wait for it to execute with some space to prove
		// it ran more times than expected:
		time.Sleep(6000 * time.Microsecond)
		tt.AssertEqual(t, numCalls, 2)
	})
}

package threads

import (
	"context"
	"fmt"
	"testing"
	"time"

	tt "github.com/blackpointcyber/threads/internal/testtools"
	"golang.org/x/sync/errgroup"
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

		var timeAfterArgs []time.Duration
		ctx = ContextWithTimeMock(ctx, tt.MockTimeAfter(func(triggerCh chan time.Time, waitCh chan time.Duration) {
			timeAfterArgs = append(timeAfterArgs, <-waitCh)
			cancel()
		}))

		var numCalls int
		var g errgroup.Group
		g.Go(func() error {
			return PeriodicWorker(2000*time.Microsecond, func(ctx context.Context) error {
				numCalls++
				return nil
			})(ctx)
		})

		g.Wait()

		tt.AssertEqual(t, numCalls, 1)
		tt.AssertEqual(t, timeAfterArgs, []time.Duration{
			2000 * time.Microsecond,
		})
	})

	t.Run("should run the callback periodically", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		ctx = ContextWithTimeMock(ctx, tt.MockTimeAfter(func(triggerCh chan time.Time, waitCh chan time.Duration) {
			<-waitCh
			triggerCh <- time.Now()
			<-waitCh
			cancel()
		}))

		var numCalls int
		go PeriodicWorker(time.Second, func(ctx context.Context) error {
			numCalls++
			return nil
		})(ctx)

		tt.AssertDone(t, 10*time.Millisecond, ctx.Done())
		tt.AssertEqual(t, numCalls, 2)
	})

	t.Run("should wait a custom amount of time if it returns a retryError", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var timeAfterArgs []time.Duration
		ctx = ContextWithTimeMock(ctx, tt.MockTimeAfter(func(triggerCh chan time.Time, waitCh chan time.Duration) {
			timeAfterArgs = append(timeAfterArgs, <-waitCh)
			cancel()
		}))

		go PeriodicWorker(42*time.Millisecond, func(ctx context.Context) error {
			return RetryWorkerIn(4 * time.Millisecond)
		})(ctx)

		tt.AssertDone(t, 10*time.Millisecond, ctx.Done())
		tt.AssertEqual(t, timeAfterArgs, []time.Duration{
			4 * time.Millisecond,
		})
	})

	t.Run("should change the interval if an adjustInterval error is returned", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var timeAfterArgs []time.Duration
		ctx = ContextWithTimeMock(ctx, tt.MockTimeAfter(func(triggerCh chan time.Time, waitCh chan time.Duration) {
			timeAfterArgs = append(timeAfterArgs, <-waitCh)
			triggerCh <- time.Now()
			timeAfterArgs = append(timeAfterArgs, <-waitCh)
			cancel()
		}))

		go PeriodicWorker(2*time.Millisecond, func(ctx context.Context) error {
			return AdjustInterval(4 * time.Millisecond)
		})(ctx)

		tt.AssertDone(t, 20*time.Millisecond, ctx.Done())
		tt.AssertEqual(t, timeAfterArgs, []time.Duration{
			2 * time.Millisecond,
			4 * time.Millisecond,
		})
	})
}

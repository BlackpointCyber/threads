package threads

import (
	"context"
	"fmt"
	"testing"
	"time"

	tt "github.bpcyber.com/vgarcia/threads/internal/testtools"
)

func TestNewGroup(t *testing.T) {
	ctx := context.Background()
	t.Run("should build correctly with a valid input context", func(t *testing.T) {
		g := NewGroup(ctx)
		tt.AssertNotEqual(t, g.g, nil)
		tt.AssertEqual(t, g.cancel != nil, true)
		tt.AssertNotEqual(t, g.ctx, ctx)

		g.cancel()
		tt.AssertErrContains(t, g.ctx.Err(), "context canceled")
		tt.AssertNoErr(t, ctx.Err())
	})
}

func TestGoAndWait(t *testing.T) {
	ctx := context.Background()

	t.Run("should wait for a single goroutine correctly", func(t *testing.T) {
		g := NewGroup(ctx)

		var executed bool
		g.Go(func(ctx context.Context) error {
			time.Sleep(100 * time.Microsecond)
			executed = true
			return nil
		})

		err := g.Wait()
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, executed, true)
	})

	t.Run("should wait for multiple goroutines correctly", func(t *testing.T) {
		g := NewGroup(ctx)

		executedFirst := make(chan struct{})
		var executedSecond bool
		g.Go(func(ctx context.Context) error {
			close(executedFirst)
			return nil
		})

		g.Go(func(ctx context.Context) error {
			time.Sleep(2000 * time.Microsecond)
			executedSecond = true
			return nil
		})

		<-executedFirst
		tt.AssertEqual(t, executedSecond, false)

		// Still waits even after the first exited:
		err := g.Wait()
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, executedSecond, true)
	})

	t.Run("should return the first error when multiple are returned", func(t *testing.T) {
		g := NewGroup(ctx)

		g.Go(func(ctx context.Context) error {
			return fmt.Errorf("first-error")
		})

		g.Go(func(ctx context.Context) error {
			time.Sleep(1000 * time.Microsecond)
			return fmt.Errorf("second-error")
		})

		err := g.Wait()
		tt.AssertErrContains(t, err, "first-error")
	})

	t.Run("should call cancel() when any of the goroutines return an error", func(t *testing.T) {
		g := NewGroup(context.TODO())
		g.Go(func(ctx context.Context) error {
			return fmt.Errorf("first-error")
		})

		var itsDone bool
		g.Go(func(ctx context.Context) error {
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				itsDone = true
			}
			return nil
		})

		err := g.Wait()
		tt.AssertErrContains(t, err, "first-error")
		tt.AssertEqual(t, itsDone, true)
	})
}

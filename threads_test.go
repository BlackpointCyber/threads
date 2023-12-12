package threads

import (
	"context"
	"fmt"
	"testing"
	"time"

	tt "github.com/blackpointcyber/threads/internal/testtools"
)

func TestForkAndWait(t *testing.T) {
	ctx := context.Background()

	t.Run("should run multiple workers in parallel", func(t *testing.T) {
		var numCallsWorker1, numCallsWorker2 int
		err := ForkAndWait(ctx,
			func(ctx context.Context) error {
				numCallsWorker1++
				return nil
			},
			func(ctx context.Context) error {
				numCallsWorker2++
				return nil
			},
		)
		tt.AssertNoErr(t, err)

		tt.AssertEqual(t, numCallsWorker1, 1)
		tt.AssertEqual(t, numCallsWorker2, 1)
	})

	t.Run("should should report errors correctly", func(t *testing.T) {
		err := ForkAndWait(ctx,
			func(ctx context.Context) error {
				return fmt.Errorf("fakeErrMsg")
			},
			func(ctx context.Context) error {
				return nil
			},
		)

		tt.AssertErrContains(t, err, "fakeErrMsg")
	})

	t.Run("should forward panics to the waiting process correctly", func(t *testing.T) {
		isCtxCanceled := make(chan struct{})

		var err error
		panicPayload, stackTrace := tt.PanicHandler(func() {
			err = ForkAndWait(ctx,
				// The function "panickingWorker" should appear on the stacktrace:
				panickingWorker,
				func(ctx context.Context) error {
					<-ctx.Done()
					isCtxCanceled <- struct{}{}
					return nil
				},
			)
		})

		tt.AssertNoErr(t, err)

		tt.AssertDone(t, 10*time.Millisecond, isCtxCanceled)
		tt.AssertContains(t, fmt.Sprint(panicPayload, stackTrace), "PanicHandler", "panickingWorker", "fakePanicPayload")
	})
}

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
		g := NewGroup(ctx)
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

	t.Run("should forward panics to the waiting process correctly", func(t *testing.T) {
		g := NewGroup(ctx)

		// The function "panickingWorker" should appear on the stacktrace:
		g.Go(panickingWorker)

		isCtxCanceled := make(chan struct{})
		g.Go(func(ctx context.Context) error {
			<-ctx.Done()
			isCtxCanceled <- struct{}{}
			return nil
		})

		var err error
		panicPayload, stackTrace := tt.PanicHandler(func() {
			err = g.Wait()
		})
		tt.AssertNoErr(t, err)

		tt.AssertDone(t, 10*time.Millisecond, isCtxCanceled)
		tt.AssertContains(t, fmt.Sprint(panicPayload, stackTrace), "PanicHandler", "panickingWorker", "fakePanicPayload")
	})
}

func panickingWorker(ctx context.Context) error {
	panic("fakePanicPayload")
}

func TestRestartGroup(t *testing.T) {
	ctx := context.Background()

	t.Run("should start restart the base group correctly", func(t *testing.T) {
		var receivedClosureValues []string

		g := NewGroup(ctx)
		closureVar := "initialState"
		g.Go(func(ctx context.Context) error {
			receivedClosureValues = append(receivedClosureValues, closureVar)
			if closureVar == "initialState" {
				closureVar = "changedState"
				return ErrRestartGroup
			}
			return nil
		})

		err := g.Wait()
		tt.AssertNoErr(t, err)

		tt.AssertEqual(t, receivedClosureValues, []string{"initialState", "changedState"})
	})

	t.Run("should create and restart subgroups correctly", func(t *testing.T) {
		var receivedClosureValues []string

		g := NewGroup(ctx)
		count := 0
		// This one should not restart:
		g.Go(func(ctx context.Context) error {
			count++
			return nil
		})

		closureVar := "initialState"
		g.SubGroup(func(ctx context.Context) error {
			receivedClosureValues = append(receivedClosureValues, closureVar)
			if closureVar == "initialState" {
				closureVar = "changedState"
				return ErrRestartGroup
			}
			return nil
		})

		err := g.Wait()
		tt.AssertNoErr(t, err)

		tt.AssertEqual(t, count, 1)
		tt.AssertEqual(t, receivedClosureValues, []string{"initialState", "changedState"})
	})
}

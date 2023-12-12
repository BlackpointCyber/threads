package main

import (
	"context"
	"fmt"
	"time"

	"github.com/blackpointcyber/threads"
)

func main() {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx = threads.ContextWithTimeAfterMock(ctx, func(triggerCh chan time.Time, waitCh chan time.Duration) {
		<-waitCh                 // Wait until time.After is called
		triggerCh <- time.Time{} // Makes <-time.After return
		<-waitCh                 // Waits again
		cancel()                 // Forces the worker to stop:
	})

	g := threads.NewGroup(ctx)

	// A worker that runs immediately at start and
	// then once again when triggerCh receives a message:
	count := 0
	g.Go(threads.PeriodicWorker(1*time.Hour, func(ctx context.Context) error {
		count++
		fmt.Printf("Run count: %v\n", count)
		return nil
	}))

	g.Wait()
}

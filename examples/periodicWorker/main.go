package main

import (
	"context"
	"fmt"
	"time"

	"github.com/blackpointcyber/threads"
)

func main() {
	ctx := context.Background()

	g := threads.NewGroup(ctx)

	// A worker that runs immediately at start and then once every second:
	g.Go(threads.PeriodicWorker(1*time.Second, func(ctx context.Context) error {
		fmt.Println("one second has passed: %v", time.Now())
		return nil
	}))

	g.Wait()
}

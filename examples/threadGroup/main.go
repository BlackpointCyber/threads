package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/blackpointcyber/threads"
)

func main() {
	ctx := context.Background()

	g := threads.NewGroup(ctx)

	g.Go(func(ctx context.Context) error {
		return DoSomeTask(ctx)
	})

	g.Go(threads.PeriodicWorker(2*time.Second, func(ctx context.Context) error {
		fmt.Println("every 2 seconds this message will show up")
		return nil
	}))

	err := g.Wait()
	if err != nil {
		log.Fatalf("unexpected error: %s", err)
	}
}

func DoSomeTask(ctx context.Context) error {
	fmt.Println("doing some task...")
	time.Sleep(3 * time.Second)
	fmt.Println("task done...")
	return nil
}

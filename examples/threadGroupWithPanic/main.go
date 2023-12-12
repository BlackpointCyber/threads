package main

import (
	"context"
	"fmt"

	"github.com/blackpointcyber/threads"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic forwarded to main:", r)
		}
	}()

	ctx := context.Background()

	g := threads.NewGroup(ctx)

	g.Go(func(ctx context.Context) error {
		panic("foo")
	})

	g.Go(func(ctx context.Context) error {
		<-ctx.Done()
		fmt.Println("context canceled because of the panic on the other Goroutine")
		return nil
	})

	// A panic on any of the Goroutines will cause g.Wait() to panic
	// immediately without waiting for the remaining Goroutines, but they
	// will still receive a cancel signal so they can make a graceful
	// shutdown.
	g.Wait()

	fmt.Println("not going to print this because of the panic")
}

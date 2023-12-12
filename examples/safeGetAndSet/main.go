package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/blackpointcyber/threads"
	"github.com/blackpointcyber/threads/safe"
)

func main() {
	ctx := context.Background()

	mux := &sync.Mutex{}

	bigArray0 := [2048]int{}

	bigArray1 := [2048]int{}
	for i := 0; i < 2048; i++ {
		bigArray1[i] = 1
	}

	g := threads.NewGroup(ctx)

	g.Go(func(ctx context.Context) error {
		bigArray1 := [2048]int{}
		for i := 0; i < 2048; i++ {
			bigArray1[i] = 1
			v := safe.Get(mux, &shared)
			safe.Set(mux, &shared, v+1)
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		for i := 0; i < 1000; i++ {
			v := safe.Get(mux, &shared)
			safe.Set(mux, &shared, v+1)
		}
		return nil
	})

	g.Wait()

	fmt.Printf("shared should be 2000 and it is in fact is: %v", shared)
}

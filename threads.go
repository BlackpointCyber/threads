package threads

import (
	"context"
	"fmt"

	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

var ErrStartGracefulShutdown = fmt.Errorf("signal to stop the execution gracefully")

type Worker func(ctx context.Context) error

type Group struct {
	g      *errgroup.Group
	ctx    context.Context
	cancel func()

	hasWaiter *atomic.Bool
	panicCh   chan any
}

func NewGroup(ctx context.Context) Group {
	ctx, cancel := context.WithCancel(ctx)

	return Group{
		g:         &errgroup.Group{},
		ctx:       ctx,
		cancel:    cancel,
		hasWaiter: &atomic.Bool{},
		panicCh:   make(chan any),
	}
}

func (g Group) Go(fn Worker) {
	g.g.Go(func() error {
		defer func() {
			if r := recover(); r != nil {
				g.cancel()
				if g.hasWaiter.Load() {
					g.panicCh <- r
					return
				}
				panic(r)
			}
		}()

		err := fn(g.ctx)
		if err != nil {
			g.cancel()
		}
		if err == ErrStartGracefulShutdown {
			return nil
		}
		return err
	})
}

func (g Group) Wait() error {
	g.hasWaiter.Store(true)
	// Set `hasWaiter` to false so that a follow up panic
	// will not send the panicPayload to a channel
	// no one is listening to:
	defer g.hasWaiter.Store(false)

	select {
	case err := <-g.waitCh():
		return err
	case panicPayload := <-g.panicCh:
		panic(panicPayload)
	}
}

func (g Group) waitCh() chan error {
	waitCh := make(chan error, 1)
	go func() {
		waitCh <- g.g.Wait()
	}()

	return waitCh
}

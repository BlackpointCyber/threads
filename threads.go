package threads

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

var ErrStartGracefulShutdown = fmt.Errorf("signal to stop the execution gracefully")
var ErrRestartGroup = fmt.Errorf("signal to restart the current threads.Group")

type Worker func(ctx context.Context) error

func ForkAndWait(ctx context.Context, fns ...Worker) error {
	g := NewGroup(ctx)
	for _, fn := range fns {
		g.Go(fn)
	}

	return g.Wait()
}

type Group struct {
	g         *errgroup.Group
	ctx       context.Context
	parentCtx context.Context
	cancel    func()

	// A list of workers to restart if requested:
	workers []Worker

	hasWaiter *atomic.Bool
	panicCh   chan any
}

func NewGroup(parentCtx context.Context) Group {
	ctx, cancel := context.WithCancel(parentCtx)

	return Group{
		g:         &errgroup.Group{},
		ctx:       ctx,
		parentCtx: parentCtx,
		cancel:    cancel,
		hasWaiter: &atomic.Bool{},
		panicCh:   make(chan any),
	}
}

func (g *Group) Go(fn Worker) {
	g.workers = append(g.workers, fn)

	g.start(fn)
}

func (g Group) start(fn Worker) {
	g.g.Go(func() error {
		defer func() {
			if r := recover(); r != nil {
				g.cancel()
				if g.hasWaiter.Load() {
					r = fmt.Sprintf("%v\n%s", r, string(debug.Stack()))
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

func (g *Group) Wait() error {
	defer func() {
		g.resetGroup()
		g.workers = []Worker{}
	}()
	g.hasWaiter.Store(true)

restartTag:
	select {
	case err := <-g.waitCh():
		if errors.Is(err, ErrRestartGroup) {
			g.resetGroup()

			for _, worker := range g.workers {
				g.start(worker)
			}

			goto restartTag
		}

		return err
	case panicPayload := <-g.panicCh:
		panic(panicPayload)
	}
}

func (g *Group) resetGroup() {
	g.g = &errgroup.Group{}
	g.ctx, g.cancel = context.WithCancel(g.parentCtx)
}

func (g *Group) SubGroup(workers ...Worker) {
	g.Go(func(ctx context.Context) error {
		subg := NewGroup(ctx)
		for _, worker := range workers {
			subg.Go(worker)
		}
		return subg.Wait()
	})
}

func (g Group) waitCh() chan error {
	waitCh := make(chan error, 1)
	go func() {
		waitCh <- g.g.Wait()
	}()

	return waitCh
}

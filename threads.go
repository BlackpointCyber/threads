package threads

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

var ErrStartGracefulShutdown = fmt.Errorf("signal to stop the execution gracefully")

type Worker func(ctx context.Context) error

type Group struct {
	g      *errgroup.Group
	ctx    context.Context
	cancel func()
}

func NewGroup(ctx context.Context) Group {
	ctx, cancel := context.WithCancel(ctx)

	return Group{
		g:      &errgroup.Group{},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (g Group) Go(fn Worker) {
	g.g.Go(func() error {
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
	return g.g.Wait()
}

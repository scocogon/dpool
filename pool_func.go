package dpool

import (
	"context"
)

type PoolFunc interface {
	ibase

	Submit(func()) func()
	SubmitContext(ctx context.Context, fn func(ctx context.Context)) func()
}

type dpFunc struct {
	*base
}

func NewPoolFunc(size int, opts ...FncOption) PoolFunc {
	return &dpFunc{
		base: newbase(size, opts...),
	}
}

func (p *dpFunc) Submit(fn func()) func() {
	return p.SubmitContext(context.Background(), func(context.Context) { fn() })
}

func (p *dpFunc) SubmitContext(ctx context.Context, fn func(ctx context.Context)) func() {
	if !p.runIf() {
		return nil
	}

	cap := int(p.Cap())

	p.addWait(cap)
	p.initContext(ctx)
	for i := 0; i < cap; i++ {
		go p.goFunc(p.ctx, fn)
	}

	return p.cancel
}

func (p *dpFunc) goFunc(ctx context.Context, fn func(context.Context)) {
	ch := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			p.doneWait()
		case <-ch:
			p.doneWait()
		}
	}()

	fn(ctx)
	close(ch)
}

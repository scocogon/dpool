package dpool

import "context"

type dpFunc struct {
	*pool

	ch chan struct{}
}

func NewPoolFunc(size int, opts ...FncOption) Pool {
	return &dpFunc{
		pool: newPool(size, opts...),
		ch:   make(chan struct{}),
	}
}

func (p *dpFunc) Submit(fn func()) {
	p.SubmitContext(context.Background(), func(context.Context) { fn() })
}

func (p *dpFunc) SubmitContext(ctx context.Context, fn func(ctx context.Context)) {
	if !p.runIf() {
		return
	}

	cap := int(p.Cap())

	p.add(cap)
	p.context(ctx)
	for i := 0; i < cap; i++ {
		go p.goFunc(p.ctx, fn)
	}
}

func (p *dpFunc) goFunc(ctx context.Context, fn func(context.Context)) {
	ch := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			p.done()
		case <-ch:
			p.done()
		}
	}()

	fn(ctx)
	close(ch)
}

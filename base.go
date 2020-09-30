package dpool

import (
	"context"
	"sync"
	"sync/atomic"
)

type ibase interface {
	Wait()
	Stop()

	addWait(int)
	doneWait()
	context() context.Context
}

type base struct {
	ctx    context.Context
	cancel func()

	wg sync.WaitGroup

	siz int32
	cap int32

	running int32

	opt *Options
}

func newbase(size int, opts ...FncOption) *base {
	if size <= 0 {
		panic("no support")
	}

	opt := defaultOptions()
	for _, f := range opts {
		f(opt)
	}

	return &base{
		cap: int32(size),
		opt: opt,
	}
}

func (p *base) addWait(size int)          { p.wg.Add(size) }
func (p *base) doneWait()                 { p.wg.Done() }
func (p *base) Wait()                     { p.wg.Wait() }
func (p *base) Size() int32               { return atomic.LoadInt32(&p.siz) }
func (p *base) Cap() int32                { return atomic.LoadInt32(&p.cap) }
func (p *base) addGR(size int32) int32    { return atomic.AddInt32(&p.siz, size) }
func (p *base) reduceGR(size int32) int32 { return atomic.AddInt32(&p.siz, -size) }
func (p *base) isRunning() bool           { return atomic.LoadInt32(&p.running) == 1 }
func (p *base) runIf() {
	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		panic("the pool has running")
	}
}
func (p *base) context() context.Context { return p.ctx }
func (p *base) initContext(ctx context.Context) {
	if p.opt.Timeout != 0 {
		p.ctx, p.cancel = context.WithTimeout(ctx, p.opt.Timeout)
	} else {
		p.ctx, p.cancel = context.WithCancel(ctx)
	}
}
func (p *base) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}
func (p *base) expansion() int32 {
	size := p.Size()
	cap := p.Cap()

	if size == 0 {
		return 1
	}

	if size >= cap {
		if !p.opt.CanAutomaticExpansion {
			return 0
		}

		if p.opt.MaxCapacity == 0 {
			return 50 // 允许自动扩容时，增加50协程
		}

		s := p.opt.MaxCapacity - size
		if s > 50 {
			return 50
		}

		return s
	}

	if cap-size >= size {
		return size
	}

	return cap - size
}
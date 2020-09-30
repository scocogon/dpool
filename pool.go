package dpool

import (
	"context"
	"sync"
	"sync/atomic"
)

type Pool interface {
	Submit(func()) (cancel func())
	SubmitContext(ctx context.Context, fn func(ctx context.Context)) (cancel func())

	Call(func(interface{})) (cancel func())
	CallContext(context.Context, func(context.Context, interface{})) (cancel func())

	Wait()

	context() context.Context
	addWorker(*worker)
	done()
}

type pool struct {
	ctx    context.Context
	cancel func()

	wg sync.WaitGroup

	siz int32
	cap int32

	running int32

	noArg bool // 是否有参数
	fn    func(context.Context, interface{})

	opt *Options
}

func newPool(size int, opts ...FncOption) *pool {
	if size <= 0 {
		panic("no support")
	}

	opt := defaultOptions()
	for _, f := range opts {
		f(opt)
	}

	return &pool{
		cap: int32(size),
		opt: opt,
	}
}

func (p *pool) Submit(func()) func()                                        { panic("No Implement") }
func (p *pool) SubmitContext(context.Context, func(context.Context)) func() { panic("No Implement") }
func (p *pool) Call(func(interface{})) func()                               { panic("No Implement") }
func (p *pool) CallContext(context.Context, func(context.Context, interface{})) func() {
	panic("No Implement")
}

func (p *pool) Release() {}
func (p *pool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}
func (p *pool) Size() int32 { return atomic.LoadInt32(&p.siz) }
func (p *pool) Cap() int32  { return atomic.LoadInt32(&p.cap) }

func (p *pool) add(size int)            { p.wg.Add(size) }
func (p *pool) done()                   { p.wg.Done() }
func (p *pool) plus(size int32) int32   { return atomic.AddInt32(&p.siz, size) }
func (p *pool) reduce(size int32) int32 { return atomic.AddInt32(&p.siz, -size) }
func (p *pool) Wait()                   { p.wg.Wait() }
func (p *pool) isRunning() bool         { return atomic.LoadInt32(&p.running) == 1 }
func (p *pool) runIf() bool {
	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		p.opt.Logger.Printf("the pool has running")
		return false
	}

	return true
}

func (p *pool) addWorker(*worker)        {}
func (p *pool) context() context.Context { return p.ctx }
func (p *pool) selfContext(ctx context.Context) {
	if p.opt.Timeout != 0 {
		p.ctx, p.cancel = context.WithTimeout(ctx, p.opt.Timeout)
	} else {
		p.ctx, p.cancel = context.WithCancel(ctx)
	}
}

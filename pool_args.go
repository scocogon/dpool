package dpool

import (
	"context"
	"errors"
	"time"
)

type dpArgs struct {
	*base

	workers chan *worker

	hasResult bool
	fn        func(context.Context, interface{}) interface{}
}

func NewPoolArgs(size int, opts ...FncOption) Pool {
	return &dpArgs{
		base:    newbase(size, opts...),
		workers: make(chan *worker, size),
	}
}

func (p *dpArgs) Call(fn func(interface{})) func() {
	return p.CallContext(context.Background(), func(_ context.Context, arg interface{}) {
		fn(arg)
	})
}

func (p *dpArgs) CallContext(ctx context.Context, fn func(context.Context, interface{})) func() {
	p.hasResult = false
	return p.callContext(ctx, func(ctx context.Context, arg interface{}) interface{} { fn(ctx, arg); return nil })
}

func (p *dpArgs) CallResult(fn func(interface{}) interface{}) func() {
	return p.CallResultContext(context.Background(), func(_ context.Context, arg interface{}) interface{} {
		return fn(arg)
	})
}

func (p *dpArgs) CallResultContext(ctx context.Context, fn func(context.Context, interface{}) interface{}) func() {
	p.hasResult = true
	return p.callContext(ctx, fn)
}

func (p *dpArgs) callContext(ctx context.Context, fn func(context.Context, interface{}) interface{}) func() {
	p.runIf()

	p.initContext(ctx)
	p.fn = fn

	p.addWait(1)
	go p.loopExpansion()

	return p.cancel
}

var ErrExit = errors.New("Stop")
var ErrTimeout = errors.New("Timeout")

func (p *dpArgs) Invoke(arg interface{}, expire ...time.Duration) (err error, res interface{}) {
	// dlog.Printf("invoke, len: %d, arg: %+v\n", len(p.workers), arg)

	select {
	case <-p.ctx.Done():
		return ErrExit, nil
	default:
	}

	var timeout time.Duration
	if len(expire) > 0 {
		timeout = expire[0]
	} else {
		timeout = 10 * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(p.context(), timeout)
	defer cancel()

	// p.opt.Logger.Printf("[pool.args] arg = %v", arg)

	select {
	case w := <-p.workers:
		w.recv(arg)
		res = w.result()

	case <-p.ctx.Done():
		return ErrExit, nil

	case <-ctx.Done():
		return ErrTimeout, nil
	}

	return
}

func (p *dpArgs) loopExpansion() {
	cap := int(p.Cap())

	defer p.doneWait()

	go func() {
		select {
		case <-p.ctx.Done():
			p.Stop()
		}
	}()

	if !p.opt.CanAutomaticExpansion {
		p.runNumWorker(int(p.Cap()))
		return
	}

	for {
		select {
		case <-p.ctx.Done():
			return

		default:
		}

		l := len(p.workers)

		switch {
		case l > cap:
		case l < int(p.defaultGorouting):
			size := p.expansion()
			// p.opt.Logger.Printf("current work: %d, new: %d\n", l, size)
			if size == 0 {
				return
			}

			p.runNumWorker(int(size))
		}

		time.Sleep(1 * time.Millisecond)
	}
}

func (p *dpArgs) runNumWorker(size int) {
	p.addWait(size)
	for i := 0; i < size; i++ {
		w := newWorker(p, p.fn)
		go w.run()
		p.addWorker(w)
	}

	p.addGR(int32(size))
}

func (p *dpArgs) addWorker(w *worker) {
	p.workers <- w
}

func (p *dpArgs) resultIf() bool {
	return p.hasResult
}

func (p *dpArgs) Stop() {
	for w := range p.workers {
		w.stop()
		p.doneWait()
	}
}

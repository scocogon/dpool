package dpool

import "context"

type worker struct {
	p Pool

	stopCH       chan struct{}
	srcCH, dstCH chan interface{}

	fn func(context.Context, interface{}) interface{}
}

func newWorker(p Pool, fn func(context.Context, interface{}) interface{}) *worker {
	return &worker{
		p:      p,
		stopCH: make(chan struct{}),
		srcCH:  make(chan interface{}, 1),
		dstCH:  make(chan interface{}),
		fn:     fn,
	}
}

func (w *worker) run() {
	for {
		select {
		case data := <-w.srcCH:
			w.work(data)

		case <-w.stopCH:
			w.release()
			return
		}
	}
}

func (w *worker) work(data interface{}) {
	res := w.fn(w.p.context(), data)
	if w.p.resultIf() {
		w.dstCH <- res
	}

	w.p.addWorker(w)
}

func (w *worker) recv(data interface{}) {
	w.srcCH <- data
}

func (w *worker) result() interface{} {
	if w.p.resultIf() {
		return <-w.dstCH
	}

	return nil
}

func (w *worker) stop() {
	w.stopCH <- struct{}{}
}

func (w *worker) release() {
	w.p = nil
	w.fn = nil
	close(w.stopCH)
	close(w.dstCH)
	close(w.srcCH)
}

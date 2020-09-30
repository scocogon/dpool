package dpool

import "context"

type worker struct {
	p Pool

	srcCH, dstCH chan interface{}

	fn func(context.Context, interface{}) interface{}
}

func newWorker(p Pool, fn func(context.Context, interface{}) interface{}) *worker {
	return &worker{
		p:     p,
		srcCH: make(chan interface{}, 1),
		dstCH: make(chan interface{}),
		fn:    fn,
	}
}

func (w *worker) run() {
	for {
		select {
		case data := <-w.srcCH:
			w.work(data)

		case <-w.p.context().Done():
			select {
			case data := <-w.srcCH:
				w.work(data)
			default:
				w.release()
				return
			}
		}
	}
}

func (w *worker) work(data interface{}) {
	w.p.addWait(1)
	res := w.fn(w.p.context(), data)
	if w.p.resultIf() {
		w.dstCH <- res
	}
	w.p.doneWait()

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

func (w *worker) release() {
	// w.p = nil
	// w.fn = nil
	// close(w.dstCH)
	// close(w.srcCH)
}

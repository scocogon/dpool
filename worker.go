package dpool

type worker struct {
	fn func(interface{})

	ch chan interface{}
}

func newWorker(fn func(interface{})) *worker {
	return &worker{
		fn: fn,
		ch: make(chan interface{}, 1),
	}
}

func (pi *worker) run() {

}

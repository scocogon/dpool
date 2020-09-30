package dpool

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolFunc(t *testing.T) {
	var ti int32

	fn1 := func() {
		atomic.AddInt32(&ti, 1)
	}

	p := NewPoolFunc(1000)
	p.Submit(fn1)
	p.Wait()

	t.Logf("ti = %d\n", ti)

	fn2 := func(context.Context) {
		time.Sleep(3 * time.Second)
		t.Log("fn2")
	}

	p = NewPoolFunc(3, WithTimeout(1*time.Second))
	p.SubmitContext(context.Background(), fn2)
	p.Wait()

	t.Log("fn2_1 done")

	p = NewPoolFunc(3)
	cancel := p.Submit(func() { fn2(context.Background()) })
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	p.Wait()
	t.Log("fn2_2 done")
}

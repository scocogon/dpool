package dpool

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolFunc(t *testing.T) {
	var res int32

	fn1 := func() {
		atomic.AddInt32(&res, 1)
	}

	p := NewPoolFunc(1000)
	p.Call(fn1)
	p.Wait()

	t.Logf("res = %d\n", res)

	fn2 := func(context.Context) {
		time.Sleep(1 * time.Second)
		atomic.AddInt32(&res, 1)
		t.Log("fn2")
	}

	p = NewPoolFunc(3, WithTimeout(2*time.Second))
	p.CallContext(context.Background(), fn2)
	p.Wait()

	t.Logf("fn2_1 done, res = %d\n", res)

	p = NewPoolFunc(3)
	cancel := p.Call(func() { fn2(context.Background()) })
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()
	p.Wait()

	t.Logf("fn2_2 done, res = %d\n", res)
}

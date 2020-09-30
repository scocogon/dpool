package dpool

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolArgs(t *testing.T) {
	var v int32
	fn := func(i interface{}) {
		atomic.AddInt32(&v, i.(int32))
	}

	p := NewPoolArgs(100, WithTimeout(10*time.Second))
	p.Call(fn)

	for i := 0; i < 1000000; i++ {
		go p.Invoke(int32(1))
	}

	p.Wait()

	t.Logf("v = %d\n", v)
}

func TestPoolArgsResult(t *testing.T) {
	var v int32
	fn := func(i interface{}) interface{} {
		time.Sleep(100 * time.Millisecond)
		return i
	}

	p := NewPoolArgs(100, WithCanAutomaticExpansion(true))
	cancel := p.CallResult(fn)

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	var f int32
	var wg sync.WaitGroup
	for i := 0; i < 8000000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err, res := p.Invoke(int32(1))
			if err != nil {
				// t.Logf("err = %s\n", err)
				atomic.AddInt32(&f, 1)
				return
			}

			atomic.AddInt32(&v, res.(int32))
		}()
	}

	p.Wait()
	wg.Wait()

	t.Logf("v = %d, failed = %d, sum = %d\n", v, f, v+f)
	t.Logf("cnt = %d\n", cnt)
}

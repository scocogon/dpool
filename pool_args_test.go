package dpool

import (
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

	p := NewPoolArgs(100)
	cancel := p.CallResult(fn)

	for i := 0; i < 1000000; i++ {
		go func() {
			err, res := p.Invoke(int32(1))
			if err != nil {
				t.Logf("err = %s\n", err)
				return
			}

			atomic.AddInt32(&v, res.(int32))
		}()
	}

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	p.Wait()

	t.Logf("v = %d\n", v)
}

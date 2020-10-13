package dpool

import (
	"context"
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

	dlog.Printf("start")
	p := NewPoolArgs(100, WithTimeout(1*time.Second))
	p.Call(fn)

	for i := 0; i < 10000000; i++ {
		go p.Invoke(int32(1))
	}

	p.Wait()

	dlog.Printf("v = %d\n", v)
}

func TestPoolArgsResult(t *testing.T) {
	var v int32
	fn := func(i interface{}) interface{} {
		// time.Sleep(100 * time.Millisecond)
		return i
	}

	p := NewPoolArgs(100, WithCanAutomaticExpansion(false))
	cancel := p.CallResult(fn)

	var f int32
	var wg sync.WaitGroup

	go func() {
		time.Sleep(4 * time.Second)
		cancel()
	}()

	for i := 0; i < 8000000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			res, err := p.Invoke(int32(1))
			if err != nil {
				atomic.AddInt32(&f, 1)
				return
			}

			atomic.AddInt32(&v, res.(int32))
		}()
	}

	p.Wait()
	wg.Wait()

	dlog.Printf("v = %d, failed = %d, sum = %d\n", v, f, v+f)
}

func BenchmarkArgsResult(b *testing.B) {
	v := int32(0)

	fn := func(ctx context.Context, i interface{}) interface{} {
		return i.(int32) * 2
	}

	// b.N = 100000
	b.ResetTimer()
	b.StartTimer()
	p := NewPoolArgs(1000)
	cancel := p.CallResultContext(context.Background(), fn)
	for i := 0; i < b.N; i++ {
		go func() {
			r, err := p.Invoke(int32(1))
			if err == nil {
				atomic.AddInt32(&v, r.(int32))
			}
		}()
	}
	cancel()
	p.Wait()
	b.StopTimer()

	// b.ResetTimer()
	// b.StartTimer()
	// wg := sync.WaitGroup{}
	// for i := 0; i < b.N; i++ {
	// 	wg.Add(1)
	// 	go func() {
	// 		atomic.AddInt32(&v, fn(context.Background(), int32(1)).(int32))
	// 		wg.Done()
	// 	}()
	// }

	// wg.Wait()
	// b.StopTimer()
}

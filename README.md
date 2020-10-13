# dpool

# 安装

```
go get -u github.com/scocogon/dpool
```

# 使用

## 执行普通函数

### 不带参数的函数

```
	ch := make(chan struct{})
	fn := func() {
		i := 0
		for {
			select {
			case <-ch: // 退出
				return

			default:
			}

			fmt.Println(i)
			i++

			time.Sleep(1 * time.Second)
		}
	}

	p := NewPoolFunc(2)
	p.Call(fn)
	time.Sleep(3 * time.Second)
	close(ch) // 协程退出

	p.Wait() // 协程池退出
```

### 带 context.Context 参数的函数

```
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_ = cancel

	fn := func(ctx0 context.Context) {
		i := 0
		for {
			select {
			case <-ctx0.Done(): // 协程池根据外部传入的 ctx0 控制退出
				return

			default:
			}

			fmt.Println(i)
			i++

			time.Sleep(1 * time.Second)
		}
	}

	p := NewPoolFunc(2)

	p.CallContext(ctx, fn)

	// 由最顶层的 cancel 退出，多用于进程退出时
	// time.Sleep(3 * time.Second)
	// cancel()

	// 仅退出协程池
	// cancel = p.CallContext(ctx, fn)
	// time.Sleep(3 * time.Second)
	// cancel()

	p.Wait()
```

## 执行带参数的函数

### 不关心返回值

```
	var res int32

	type pas struct {
		ctx context.Context
		i   int32
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	fn := func(ctx0 context.Context, param interface{}) {
		val := param.(*pas)

		select {
		// case <-ctx0.Done(): // 协程池退出，正常来说数据处理函数是不需要关心协程池是否结束的
		// 	return

		case <-val.ctx.Done(): // 数据退出
			return

		default:
		}

		atomic.AddInt32(&res, val.i)
	}

	p := NewPoolArgs(10)
	p.CallContext(ctx, fn)

	_ = cancel

	// 由最顶层的 cancel 退出，多用于进程退出时
	// time.Sleep(3 * time.Second)
	// cancel()

	// 仅停止协程池
	// cancel = p.CallContext(ctx, fn)
	// time.Sleep(3 * time.Second)
	// cancel()

	for i := 0; i < 1000; i++ {
		go p.Invoke(&pas{ctx, int32(1)}) // 传参
	}

	p.Wait()

	fmt.Println("done, res =", res)
```

### 关心数据处理的返回值

```
	var res int32

	type pas struct {
		ctx context.Context
		i   int32
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	fn := func(ctx0 context.Context, param interface{}) interface{} {
		val := param.(*pas)

		select {
		// case <-ctx0.Done(): // 协程池退出，正常来说数据处理函数是不需要关心协程池是否结束的
		// 	return

		case <-val.ctx.Done(): // 数据退出
			return nil

		default:
		}

		return val.i
	}

	p := NewPoolArgs(10)
	p.CallResultContext(ctx, fn)

	_ = cancel

	// 由最顶层的 cancel 退出，多用于进程退出时
	// time.Sleep(3 * time.Second)
	// cancel()

	// 仅停止协程池
	// cancel = p.CallResultContext(ctx, fn)
	// time.Sleep(3 * time.Second)
	// cancel()

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			r, err := p.Invoke(&pas{ctx, int32(1)}) // 传参
			if err != nil {
				t.Errorf("Invode err: %s\n", err) // err = Timeout 表示协程池小了，数据处理不过来
				return
			}

			atomic.AddInt32(&res, r.(int32))
			wg.Done()
		}()
	}

	p.Wait()
	wg.Wait()

	fmt.Println("done, res =", res)
```
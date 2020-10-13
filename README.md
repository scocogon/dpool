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

```

```
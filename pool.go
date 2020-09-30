package dpool

import (
	"context"
	"time"
)

type Pool interface {
	ibase

	Call(func(interface{})) (cancel func())
	CallContext(context.Context, func(context.Context, interface{})) (cancel func())

	CallResult(fn func(interface{}) interface{}) func()
	CallResultContext(ctx context.Context, fn func(context.Context, interface{}) interface{}) func()

	Invoke(interface{}, ...time.Duration) (err error, res interface{})

	addWorker(*worker)
	resultIf() bool
}

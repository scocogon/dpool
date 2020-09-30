package dpool

import (
	"time"
)

type FncOption func(opt *Options)

type Options struct {
	// 允许自动扩容，默认禁止
	CanAutomaticExpansion bool

	// 自动扩容的最大容量
	MaxCapacity int32

	// 是否异步，默认异步
	NonBlock bool

	// 超时
	Timeout time.Duration

	// Logger
	Logger Logger
}

func defaultOptions() *Options {
	return &Options{
		CanAutomaticExpansion: false,
		MaxCapacity:           100000,
		NonBlock:              true,
		Timeout:               0,
		Logger:                dlog,
	}
}

func WithCanAutomaticExpansion(CanAutomaticExpansion bool) FncOption {
	return func(opt *Options) {
		opt.CanAutomaticExpansion = CanAutomaticExpansion
	}
}
func WithMaxCapacity(MaxCapacity int32) FncOption {
	return func(opt *Options) {
		opt.MaxCapacity = MaxCapacity
	}
}
func WithNonBlock(NonBlock bool) FncOption {
	return func(opt *Options) {
		opt.NonBlock = NonBlock
	}
}
func WithTimeout(Timeout time.Duration) FncOption {
	return func(opt *Options) {
		opt.Timeout = Timeout
	}
}
func WithLogger(Logger Logger) FncOption {
	return func(opt *Options) {
		opt.Logger = Logger
	}
}

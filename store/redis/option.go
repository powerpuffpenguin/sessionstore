package redis

import "context"

var defaultOptions = options{
	ctx: context.Background(),
}

type options struct {
	ctx   context.Context
	write Backend
	read  Backend
}
type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (fdo *funcOption) apply(do *options) {
	fdo.f(do)
}
func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}
func WithRead(read Redis) Option {
	return newFuncOption(func(o *options) {
		o.read = read
	})
}
func WithContext(ctx context.Context) Option {
	return newFuncOption(func(o *options) {
		if ctx == nil {
			o.ctx = context.Background()
		} else {
			o.ctx = ctx
		}
	})
}

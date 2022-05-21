package sessionstore

import (
	"time"

	"github.com/powerpuffpenguin/sessionstore/cryptoer"
)

var defaultOptions = options{
	method:   cryptoer.SigningMethodHMD5,
	key:      []byte(`cerberu is an idea`),
	access:   time.Hour,
	refresh:  time.Hour * 12 * 3,
	deadline: time.Hour * 24 * 30,
}

type options struct {
	// 簽名算法
	method cryptoer.SigningMethod
	key    []byte
	// token 有效期
	access  time.Duration
	refresh time.Duration
	// token 最長可維持多久(一直 refresh 最長 session 時長)，如果爲 0 則不限制
	deadline time.Duration

	// 存儲後端
	store Store
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
func WithMethod(method cryptoer.SigningMethod) Option {
	return newFuncOption(func(o *options) {
		o.method = method
	})
}
func WithKey(key []byte) Option {
	return newFuncOption(func(o *options) {
		o.key = key
	})
}
func WithAccess(access time.Duration) Option {
	return newFuncOption(func(o *options) {
		o.access = access
	})
}
func WithRefresh(refresh time.Duration) Option {
	return newFuncOption(func(o *options) {
		o.refresh = refresh
	})
}
func WithDeadline(deadline time.Duration) Option {
	return newFuncOption(func(o *options) {
		if deadline < 0 {
			deadline = 0
		}
		o.deadline = deadline
	})
}
func WithStore(store Store) Option {
	return newFuncOption(func(o *options) {
		o.store = store
	})
}

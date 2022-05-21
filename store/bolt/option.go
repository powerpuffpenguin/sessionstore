package bolt

import (
	"github.com/boltdb/bolt"
)

var defaultOptions = options{
	url:   `bolt://0600/bbolt.db?Timeout=1s`,
	limit: 1000 * 1000 * 10,
}

type options struct {
	db    *bolt.DB
	url   string
	limit int64
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
func WithDB(db *bolt.DB) Option {
	return newFuncOption(func(o *options) {
		o.db = db
	})
}
func WithURL(url string) Option {
	return newFuncOption(func(o *options) {
		o.url = url
	})
}
func WithLimit(limit int64) Option {
	return newFuncOption(func(o *options) {
		if limit < 1 {
			o.limit = 1000 * 1000 * 10
		} else {
			o.limit = limit
		}
	})
}

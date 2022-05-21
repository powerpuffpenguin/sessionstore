package redis

import (
	"context"
	"errors"
	"time"
)

type Store struct {
	opts *options
}

func New(redis Backend, opt ...Option) (s *Store, e error) {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	if redis == nil {
		e = errors.New(`redis not supported nil`)
		return
	}
	opts.write = redis
	if opts.read == nil {
		opts.read = redis
	}
	s = &Store{
		opts: &opts,
	}
	return
}

// 設置數據
func (s *Store) Put(ctx context.Context, key string, value []byte, deadline time.Time) (e error) {
	expiration := time.Until(deadline)
	if expiration >= time.Second {
		e = s.opts.write.Set(ctx, key, value, expiration).Err()
	}
	return
}

// 返回數據
func (s *Store) Get(ctx context.Context, key string) (value []byte, e error) {
	return s.opts.read.Get(ctx, key).Bytes()
}

// 刪除數據
func (s *Store) Del(ctx context.Context, key string) (e error) {
	return s.opts.write.Del(ctx, key).Err()
}

// 刪除指定前綴的數據
func (s *Store) DelPrefix(ctx context.Context, prefix string) (e error){
	var (
		cursor uint64
		count  int64 = 1000
		match        = prefix + "*"
	)
	for {
		scan := s.opts.write.Scan(s.opts.ctx, cursor,
			match,
			count,
		)
		var (
			keys []string
			e    error
		)
		keys, cursor, e = scan.Result()
		if e != nil {
			return e
		}
		if len(keys) != 0 {
			e = s.opts.write.Del(s.opts.ctx, keys...).Err()
			if e != nil {
				return e
			}
		}
		if cursor == 0 {
			break
		}
	}
	return
}
// 關閉存儲設備 釋放相關資源
func (s *Store) Close() (e error) {
	e = s.opts.write.Close()
	if s.opts.read != s.opts.write {
		s.opts.read.Close()
	}
	return
}

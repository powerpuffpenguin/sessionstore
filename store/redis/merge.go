package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type _MergeArgs struct {
	run    func(pipeliner redis.Pipeliner) interface{}
	done   func(v interface{})
	result interface{}
}
type Merge struct {
	redis  Redis
	count  int
	ch     chan *_MergeArgs
	ctx    context.Context
	cancel context.CancelFunc
}

func NewMerge(ctx context.Context, redis Redis, count int) *Merge {
	if count < 5 {
		count = 5
	}
	ctx, cancel := context.WithCancel(ctx)
	m := &Merge{
		redis:  redis,
		count:  count,
		ch:     make(chan *_MergeArgs),
		ctx:    ctx,
		cancel: cancel,
	}
	go m.run()
	return m
}
func (m *Merge) Close() error {
	m.cancel()
	return m.redis.Close()
}

func (m *Merge) run() {
	var (
		result = make([]*_MergeArgs, 0, m.count)
		args   *_MergeArgs
		done   = m.ctx.Done()
	)
	for {
		select {
		case args = <-m.ch:
		case <-done:
			return
		}
		pipeliner := m.redis.Pipeline()
		args.result = args.run(pipeliner)
		result = append(result, args)

	Merge:
		for len(result) != m.count {
			select {
			case args = <-m.ch:
				args.result = args.run(pipeliner)
				result = append(result, args)
			case <-done:
				// exec
				pipeliner.Exec(context.Background())
				for _, item := range result {
					item.done(item.result)
				}
				return
			default:
				break Merge
			}
		}

		// exec
		pipeliner.Exec(context.Background())
		for _, item := range result {
			item.done(item.result)
		}

		// reset
		result = result[:0]
	}
}
func (m *Merge) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	done := make(chan *redis.StatusCmd, 1)
	select {
	case <-m.ctx.Done():
		var result redis.StatusCmd
		result.SetErr(m.ctx.Err())
		return &result
	case m.ch <- &_MergeArgs{
		run: func(pipeliner redis.Pipeliner) interface{} {
			return pipeliner.Set(ctx, key, value, expiration)
		},
		done: func(v interface{}) {
			done <- v.(*redis.StatusCmd)
		},
	}:
		return <-done
	}
}
func (m *Merge) Get(ctx context.Context, key string) *redis.StringCmd {
	done := make(chan *redis.StringCmd, 1)
	select {
	case <-m.ctx.Done():
		var result redis.StringCmd
		result.SetErr(m.ctx.Err())
		return &result
	case m.ch <- &_MergeArgs{
		run: func(pipeliner redis.Pipeliner) interface{} {
			return pipeliner.Get(ctx, key)
		},
		done: func(v interface{}) {
			done <- v.(*redis.StringCmd)
		},
	}:
		return <-done
	}
}
func (m *Merge) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	done := make(chan *redis.IntCmd, 1)
	select {
	case <-m.ctx.Done():
		var result redis.IntCmd
		result.SetErr(m.ctx.Err())
		return &result
	case m.ch <- &_MergeArgs{
		run: func(pipeliner redis.Pipeliner) interface{} {
			return pipeliner.Del(ctx, keys...)
		},
		done: func(v interface{}) {
			done <- v.(*redis.IntCmd)
		},
	}:
		return <-done
	}
}
func (m *Merge) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	done := make(chan *redis.ScanCmd, 1)
	select {
	case <-m.ctx.Done():
		var result redis.ScanCmd
		result.SetErr(m.ctx.Err())
		return &result
	case m.ch <- &_MergeArgs{
		run: func(pipeliner redis.Pipeliner) interface{} {
			return pipeliner.Scan(ctx, cursor, match, count)
		},
		done: func(v interface{}) {
			done <- v.(*redis.ScanCmd)
		},
	}:
		return <-done
	}
}

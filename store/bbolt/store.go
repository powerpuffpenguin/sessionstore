package bbolt

import (
	"context"
	"time"

	"github.com/powerpuffpenguin/sessionstore"
	"github.com/powerpuffpenguin/sessionstore/store"
	bolt "go.etcd.io/bbolt"
)

type Store struct {
	opts *options
}

func New(opt ...Option) (s *Store, e error) {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	if opts.db == nil {
		path, mode, options, err := ParseURL(opts.url)
		if err != nil {
			e = err
			return
		}
		db, err := bolt.Open(path, mode, options)
		if err != nil {
			e = err
			return
		}
		opts.db = db
	}
	s = &Store{
		opts: &opts,
	}
	return
}

// 設置數據
func (s *Store) Put(ctx context.Context, key string, value []byte, deadline time.Time) (err error) {
	expiration := time.Until(deadline)
	if expiration < time.Second {
		return
	}
	err = s.opts.db.Update(func(t *bolt.Tx) (e error) {
		bkey := sessionstore.StringToBytes(key)

		bsystem, bdata, bsort := getBuckets(t)
		if bdata == nil || bsort == nil || bsystem == nil {
			bsystem, bdata, bsort, e = createBuckets(t)
			if e != nil {
				return
			}
		} else {
			e = delKey(bsystem, bdata, bsort, bkey)
			if e != nil {
				return
			}
			e = popKey(bsystem, bdata, bsort)
			if e != nil {
				return
			}
		}
		count := getCount(bsystem)
		if count >= s.opts.limit {
			e = store.ErrCapacityLimitReached
			return
		}

		unix := deadline.Unix()

		id, e := putSort(bsort, bkey, unix)
		if e != nil {
			return
		}
		e = putData(bdata, id, bkey, value, unix)
		if e != nil {
			return
		}

		e = setCount(bsystem, count+1)
		return
	})
	return
}

// 返回數據
func (s *Store) Get(ctx context.Context, key string) (value []byte, err error) {
	err = s.opts.db.View(func(t *bolt.Tx) (e error) {
		bdata := t.Bucket(bucketData)
		if bdata == nil {
			return
		}
		bkey := sessionstore.StringToBytes(key)
		data, e := getData(bdata, bkey)
		if e != nil || data == nil {
			return
		}
		if time.Now().Unix() > data.Deadline {
			return
		}
		value = data.Data
		return
	})
	return
}

// 刪除數據
func (s *Store) Del(ctx context.Context, key string) error {
	return s.opts.db.Update(func(t *bolt.Tx) (e error) {
		bsystem, bdata, bsort := getBuckets(t)
		if bsystem == nil || bdata == nil || bsort == nil {
			return
		}
		bkey := sessionstore.StringToBytes(key)
		e = delKey(bsystem, bdata, bsort, bkey)
		return
	})
}

// 刪除指定前綴的數據
func (s *Store) DelPrefix(ctx context.Context, prefix string) (err error) {
	err = s.opts.db.Update(func(t *bolt.Tx) (e error) {
		bsystem, bdata, bsort := getBuckets(t)
		if bdata == nil || bsort == nil {
			return
		}
		e = delKeyPrefix(bsystem, bdata, bsort, prefix)
		return
	})
	return
}

func (s *Store) Close() (e error) {
	e = s.opts.db.Close()
	return
}

// 清空所有 session
func (s *Store) Reset() (err error) {
	err = s.opts.db.Update(func(t *bolt.Tx) (e error) {
		names := [][]byte{
			bucketSystem, bucketData, bucketSort,
		}
		for _, name := range names {
			if t.Bucket(name) != nil {
				e = t.DeleteBucket(name)
				if e != nil {
					return
				}
			}
		}
		return
	})
	return
}

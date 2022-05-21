package sessionstore

import (
	"context"
	"time"
)

type Store interface {
	// 設置數據
	Put(ctx context.Context, key string, value []byte, deadline time.Time) (e error)
	// 返回數據
	Get(ctx context.Context, key string) (value []byte, e error)
	// 刪除數據
	Del(ctx context.Context, key string) (e error)
	// 刪除指定前綴的數據
	DelPrefix(ctx context.Context, prefix string) (e error)
	// 關閉存儲設備 釋放相關資源
	Close() (e error)
}

# sessionstore

對於 http 等無狀態協議，需要使用 session 來保持會話，通常如果不是 JWT 服務器需要將 session 存儲起來，本庫的目的就在於抽象化存儲方案，只要實現 Store 接口即可

```
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
```
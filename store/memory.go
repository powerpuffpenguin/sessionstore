package store

import (
	"container/list"
	"context"
	"strings"
	"sync"
	"time"
)

type _MemoryValue struct {
	key      string
	value    []byte
	deadline time.Time
}

func (v *_MemoryValue) IsDeleted() bool {
	return time.Now().After(v.deadline)
}

type Memory struct {
	maxsize int
	keys    map[string]*list.Element
	list    *list.List
	rw      sync.RWMutex
	closed  bool
	done    chan struct{}
}

func NewMemory(maxsize int) *Memory {
	if maxsize < 1 {
		maxsize = 1
	}
	m := &Memory{
		keys:    make(map[string]*list.Element),
		list:    list.New(),
		maxsize: maxsize,
		done:    make(chan struct{}),
	}
	go m.clearWorker()
	return m
}
func (m *Memory) clearWorker() {
	var t *time.Timer
	for {
		if t == nil {
			t = time.NewTimer(time.Minute * 5)
		} else {
			t.Reset(time.Minute * 5)
		}

		select {
		case <-m.done:
			if !t.Stop() {
				<-t.C
			}
			return
		case <-t.C:
		}

		m.rw.Lock()
		if m.closed {
			m.rw.Unlock()
			return
		}
		m.pop()
		m.rw.Unlock()
	}
}

// 設置數據
func (m *Memory) Put(ctx context.Context, key string, value []byte, deadline time.Time) (e error) {
	if !time.Now().Before(deadline) {
		return
	}
	m.rw.Lock()
	if m.closed {
		e = ErrClosed
	} else {
		e = m.put(key, &_MemoryValue{
			key:      key,
			value:    value,
			deadline: deadline,
		})
	}
	m.rw.Unlock()
	return
}
func (m *Memory) pop() {
	var (
		now      = time.Now()
		iterator = m.list.Front()
		next     *list.Element
	)
	for iterator != nil {
		value := iterator.Value.(*_MemoryValue)
		if now.After(value.deadline) {
			next = iterator.Next()

			m.list.Remove(iterator)
			delete(m.keys, value.key)

			iterator = next
		} else {
			break
		}
	}
}
func (m *Memory) put(key string, value *_MemoryValue) (e error) {
	m.pop()

	ele, ok := m.keys[key]
	if ok {
		ele.Value = value
		m.list.MoveToBack(ele)
		return
	}
	if m.maxsize == m.list.Len() {
		e = ErrCapacityLimitReached
		return
	}

	ele = m.list.PushBack(value)
	m.keys[key] = ele
	return
}

// 返回數據
func (m *Memory) Get(ctx context.Context, key string) (value []byte, e error) {
	m.rw.RLock()
	if m.closed {
		e = ErrClosed
	} else {
		if ele, ok := m.keys[key]; ok {
			val := ele.Value.(*_MemoryValue)
			if !val.IsDeleted() {
				value = val.value
			}
		}
	}
	m.rw.RUnlock()
	return
}

// 刪除數據
func (m *Memory) Del(ctx context.Context, key string) (e error) {
	m.rw.Lock()
	if m.closed {
		e = ErrClosed
	} else {
		if ele, ok := m.keys[key]; ok {
			delete(m.keys, key)
			m.list.Remove(ele)
		}
	}
	m.rw.Unlock()
	return
}

// 刪除指定前綴的數據
func (m *Memory) DelPrefix(ctx context.Context, prefix string) (e error) {
	m.rw.Lock()
	if m.closed {
		e = ErrClosed
	} else {
		for key, ele := range m.keys {
			if strings.HasPrefix(key, prefix) {
				delete(m.keys, key)
				m.list.Remove(ele)
			}
		}
	}
	m.rw.Unlock()
	return
}

// 關閉存儲設備 釋放相關資源
func (m *Memory) Close() (e error) {
	m.rw.Lock()
	if m.closed {
		e = ErrClosed
	} else {
		m.closed = true
		close(m.done)
	}
	m.rw.Unlock()
	return
}

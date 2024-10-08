package mempool

import (
	"fmt"
	"runtime"
)

type IQueue interface {
	Init(int64)
	Push(int64)
	Pop() (int64, bool)
}

type MemPool[T any] struct {
	queue casQueue
	cache bitmapCache[T]
}

func (m *MemPool[T]) Init(size int64) {
	m.cache.init(size)
	m.queue.Init(size)
	for i := int64(0); i < size; i++ {
		m.queue.Push(i)
	}
}

func (m *MemPool[T]) New() *T {
	idx, ok := m.queue.Pop()
	// idx, ok, w, r := m.queue.pop()
	if ok {
		if m.cache.tag[idx].Load() {

			panic(fmt.Sprintf(
				"cache[%d] not recycled",
				idx,
			))
		}
		m.cache.tag[idx].Store(true)
		return &m.cache.cache[idx]
	}
	return nil
}

func (m *MemPool[T]) Free(ptr *T) bool {
	idx := int64(m.cache.getIndex(ptr))
	if idx < m.cache.size {
		if m.cache.tag[idx].CompareAndSwap(true, false) {
			for !m.queue.Push(idx) {
				runtime.Gosched()
			}
			return true
		}
	}
	return false
}

package mempool

type chQueue struct {
	ch chan int64
}

func (q *chQueue) Init(size int64) {
	q.ch = make(chan int64, size)
}

func (q *chQueue) Push(idx int64) bool {
	select {
	case q.ch <- idx:
		return true
	default:
		return false
	}
}

func (q *chQueue) Pop() (idx int64, ok bool) {
	select {
	case idx = <-q.ch:
		return idx, true
	default:
		return -1, false
	}
}

type ChMemPool[T any] struct {
	queue chQueue
	cache bitmapCache[T]
}

func (m *ChMemPool[T]) Init(size int64) {
	m.cache.init(size)
	m.queue.Init(size)
	for i := int64(0); i < size; i++ {
		m.queue.Push(i)
	}
}

func (m *ChMemPool[T]) New() *T {
	idx, ok := m.queue.Pop()
	if ok {
		m.cache.tag[idx].Store(true)
		return &m.cache.cache[idx]
	}
	return nil
}

func (m *ChMemPool[T]) Free(ptr *T) bool {
	idx := int64(m.cache.getIndex(ptr))
	if idx < m.cache.size {
		if m.cache.tag[idx].CompareAndSwap(true, false) {
			return m.queue.Push(idx)
		}
	}
	return false
}

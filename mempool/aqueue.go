package mempool

import "sync/atomic"

// queue:  0 < w - r < = size
// [_ _ _ _ _ _ _]_ _
// [r ---------->]w
type aQueue struct {
	data []int64
	size int64
	w    int64 // write position
	cr   int64 // confirmed read position
	r    int64 // read position
	cw   int64 // confirmed write position
}

func (q *aQueue) Init(size int64) {
	q.size = size
	q.data = make([]int64, size)
	q.w = 0
	q.r = 0
	q.cr = 0
	q.cw = 0
}

// aPush
// avoid full: w - r > size
// _ _ _ _ _ _ _ _
// r ----------> w
func (q *aQueue) Push(idx int64) bool {
	w := atomic.AddInt64(&q.w, 1)
	r := atomic.LoadInt64(&q.cr)
	if w-r > q.size {
		atomic.AddInt64(&q.w, -1)
		return false
	} else {
		q.data[(w-1)%q.size] = idx
		atomic.AddInt64(&q.cw, 1)
		return true
	}
}

// aPop
// avoid empty: r = w
// _ _ _ _ _ _ _ _
// r
// w
func (q *aQueue) Pop() (int64, bool) {
	r := atomic.AddInt64(&q.r, 1) - 1
	w := atomic.LoadInt64(&q.cw)
	if r < w {
		atomic.AddInt64(&q.cr, 1)
		return q.data[r%q.size], true
	} else {
		atomic.AddInt64(&q.r, -1)
		return -1, false
	}
}

type AQueue = aQueue

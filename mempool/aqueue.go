package mempool

import (
	"sync/atomic"
)

type casData struct {
	d int64
	t atomic.Bool
}

type casQueue struct {
	data []casData
	size int64
	w    int64 // write position
	r    int64 // read position

}

func (q *casQueue) Init(size int64) {
	q.size = size
	q.data = make([]casData, size)
	q.w = 0
	q.r = 0

}

// aPush
// avoid full: w - r > size
// _ _ _ _ _ _ _ _
// r ----------> w
func (q *casQueue) Push(idx int64) bool {
	for {
		r := atomic.LoadInt64(&q.r)
		w := atomic.LoadInt64(&q.w)
		if q.notFull(w, r) {
			data := &q.data[w%q.size]
			if data.t.Load() {
				return false
			}
			if atomic.CompareAndSwapInt64(&q.w, w, w+1) {
				data.d = idx
				data.t.Store(true)
				return true
			}
		} else {
			return false
		}
	}
}

func (q *casQueue) notFull(w, r int64) bool {
	return w-r < q.size
}

// aPop
// avoid empty: r = w
// _ _ _ _ _ _ _ _
// r
// w
func (q *casQueue) Pop() (int64, bool) {
	for {
		w := atomic.LoadInt64(&q.w)
		r := atomic.LoadInt64(&q.r)
		if q.notEmpty(w, r) {
			d := &q.data[r%q.size]
			if !d.t.Load() {
				return -1, false
			}
			data := d.d
			if atomic.CompareAndSwapInt64(&q.r, r, r+1) {
				d.t.Store(false)
				return data, true
			}
		} else {
			return -1, false
		}
	}
}

func (q *casQueue) notEmpty(w, r int64) bool {
	return r < w
}

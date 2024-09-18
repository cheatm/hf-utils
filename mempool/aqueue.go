package mempool

import (
	"runtime"
	"sync/atomic"
)

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
		idx := q.data[r%q.size]
		atomic.AddInt64(&q.cr, 1)
		return idx, true
	} else {
		atomic.AddInt64(&q.r, -1)
		return -1, false
	}
}

type AQueue = aQueue

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
				runtime.Gosched()
				continue
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

const (
	UpperBit uint64 = 0xFFFFFFFF00000000
	LowerBit uint64 = 0x00000000FFFFFFFF
)

type aQueueShift32 struct {
	data []int64
	rw   uint64 // u64[r[u32]w[u32]]
	size uint32

	popFailed  int64
	pushFailed int64
}

func (a *aQueueShift32) Init(size int64) {
	a.size = uint32(size + 1)
	a.data = make([]int64, a.size)

}

// full : (w+1) % size = r
// _ _ _ _ _ _ _ _
// r ----------> w
func (a *aQueueShift32) Push(val int64) bool {

	for {

		rw := atomic.LoadUint64(&a.rw)
		r := uint32((rw & UpperBit) >> 32)
		w := uint32(rw & LowerBit)
		if (w+1)%a.size != r {
			a.data[w] = val
			w = (w + 1) % a.size
			if atomic.CompareAndSwapUint64(&a.rw, rw, uint64(w)+(uint64(r)<<32)) {
				return true
			}
			atomic.AddInt64(&a.popFailed, 1)
		} else {
			return false
		}
	}

}

// aPush
// avoid empty: w = r
// _ _ _ _ _ _ _ _
// r
// w
func (a *aQueueShift32) Pop() (int64, bool) {
	for {
		rw := atomic.LoadUint64(&a.rw)
		r := uint32((rw & UpperBit) >> 32)
		w := uint32(rw & LowerBit)
		if w != r {
			val := a.data[r]
			r = (r + 1) % a.size
			if atomic.CompareAndSwapUint64(&a.rw, rw, uint64(w)+(uint64(r)<<32)) {
				return val, true
			}
			atomic.AddInt64(&a.pushFailed, 1)
		} else {
			return -1, false
		}
	}
}

package mempool

import "sync/atomic"

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

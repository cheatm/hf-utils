package mempool

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
)

type iMemPool[T any] interface {
	Free(ptr *T) bool
	New() *T
	Init(int64)
}

type RawMemPool[T any] struct{}

func (p *RawMemPool[T]) Free(*T) bool {
	return false
}

func (p *RawMemPool[T]) New() *T {
	return new(T)
}

func (p *RawMemPool[T]) Init(int64) {}

type object struct {
	By    int64
	Data  [128]byte
	Count int64
}

func (o *object) Require() int64 {
	return atomic.AddInt64(&o.Count, 1)
}

func (o *object) Release() int64 {
	return atomic.AddInt64(&o.Count, -1)
}

type PoolTester struct {
	pool     iMemPool[object]
	size     int64
	batch    int64
	parallel int
	id       atomic.Int64
}

func (pt *PoolTester) BenchmarkRandomRW(b *testing.B) {
	var count int
	if pt.parallel > 0 {
		runtime.GOMAXPROCS(pt.parallel)
		count = pt.parallel
	} else {
		count = runtime.GOMAXPROCS(0)
	}
	ch := make(chan int64, count)
	pt.pool.Init(pt.size)
	b.RunParallel(func(pb *testing.PB) {
		allocCount := pt.BenchmarkParallel(pb)
		ch <- int64(allocCount)
	})
	results := make([]int64, count)
	totalAlloc := int64(0)
	for i := 0; i < count; i++ {
		results[i] = <-ch
		totalAlloc += results[i]
	}
	b.Logf("N=%d, AllocRate=%f", b.N, float64(totalAlloc)/float64(b.N))
}

func (pt *PoolTester) BenchmarkParallel(b *testing.PB) int {
	// pid := pt.id.Add(1)
	array := make([]*object, pt.batch)
	var allocCount int = 0
	var i int64 = 0
	for b.Next() {
		i++
		_i := i % pt.batch
		if array[_i] == nil {
			ptr := pt.pool.New()
			if ptr == nil {
				continue
			}
			required := ptr.Require()
			if required != 1 {
				panic(fmt.Errorf("Require Failed: {i=%d, _i=%d, r=%d}", i, _i, required))
			}
			// if ptr.Count != 0 {
			// 	idx := ptr.Count % pt.batch
			// 	ptr2 := array[idx]
			// 	if ptr2 != nil {
			// 		panic(fmt.Errorf(
			// 			"[pid=%d] count should be zero, not %d, array[%d]{Count=%d, By=%d}, current count: %d",
			// 			pid, ptr.Count, idx, ptr2.Count, ptr2.By, i,
			// 		))
			// 	}
			// 	panic(fmt.Errorf(
			// 		"[pid=%d] count should be zero, not %d, current count: %d",
			// 		pid, ptr.Count, i,
			// 	))
			// }
			array[_i] = ptr
			allocCount++
		} else {
			// array[_i].Count = 0
			// array[_i].By = 0
			release := array[_i].Release()
			if release != 0 {
				panic(fmt.Errorf("Release Failed: {i=%d, _i=%d, r=%d}", i, _i, release))
			}
			pt.pool.Free(array[_i])
			array[_i] = nil
		}

	}
	return allocCount
}

func BenchmarkMemPoolRW(b *testing.B) {
	pt := &PoolTester{
		pool:     &MemPool[object]{},
		size:     (1 << 24) - 1,
		batch:    1 << 12,
		parallel: 4,
	}
	pt.BenchmarkRandomRW(b)
}

func BenchmarkChMemPoolRW(b *testing.B) {
	pt := &PoolTester{
		pool:     &ChMemPool[object]{},
		size:     1 << 20,
		batch:    1 << 12,
		parallel: 4,
	}
	pt.BenchmarkRandomRW(b)
}

func BenchmarkRawPoolRW(b *testing.B) {
	pt := &PoolTester{
		pool:     &RawMemPool[object]{},
		size:     1 << 16,
		batch:    1 << 12,
		parallel: 2,
	}
	pt.BenchmarkRandomRW(b)
}

package mempool

import (

	// "math/rand"
	"runtime"
	"testing"

	"golang.org/x/exp/rand"
)

// type iMemPool[T any] interface {
// 	Free(ptr *T) bool
// 	New() *T
// 	Init(int64)
// }

// type RawMemPool[T any] struct{}

// func (p *RawMemPool[T]) Free(*T) bool {
// 	return false
// }

// func (p *RawMemPool[T]) New() *T {
// 	return new(T)
// }

// func (p *RawMemPool[T]) Init(int64) {}

// type object struct {
// 	Data [128]byte
// }

// type PoolTester struct {
// 	pool     iMemPool[object]
// 	size     int64
// 	batch    int64
// 	parallel int
// }

// func (pt *PoolTester) BenchmarkRandomRW(b *testing.B) {
// 	var count int
// 	if pt.parallel > 0 {
// 		runtime.GOMAXPROCS(pt.parallel)
// 		count = pt.parallel
// 	} else {
// 		count = runtime.GOMAXPROCS(0)
// 	}
// 	ch := make(chan int64, count)
// 	pt.pool.Init(pt.size)
// 	b.RunParallel(func(pb *testing.PB) {
// 		allocCount := pt.BenchmarkParallel(pb)
// 		ch <- int64(allocCount)
// 	})
// 	results := make([]int64, count)
// 	totalAlloc := int64(0)
// 	for i := 0; i < count; i++ {
// 		results[i] = <-ch
// 		totalAlloc += results[i]
// 	}
// 	b.Logf("N=%d, AllocRate=%f, %+v", b.N, float64(totalAlloc)/float64(b.N), results)

// }

// func (pt *PoolTester) BenchmarkParallel(b *testing.PB) int {
// 	array := make([]*object, pt.batch)
// 	var allocCount int = 0
// 	for b.Next() {
// 		i := rand.Int63n(pt.batch)
// 		if array[i] == nil {
// 			array[i] = pt.pool.New()
// 			allocCount++
// 		} else {
// 			pt.pool.Free(array[i])
// 			array[i] = nil
// 		}

// 	}
// 	return allocCount
// }

// func BenchmarkMemPoolRW(b *testing.B) {
// 	pt := &PoolTester{
// 		pool:     &MemPool[object]{},
// 		size:     1 << 16,
// 		batch:    1000,
// 		parallel: 2,
// 	}
// 	pt.BenchmarkRandomRW(b)
// }

// func BenchmarkRawPoolRW(b *testing.B) {
// 	pt := &PoolTester{
// 		pool:     &RawMemPool[object]{},
// 		size:     1 << 16,
// 		batch:    1000,
// 		parallel: 2,
// 	}
// 	pt.BenchmarkRandomRW(b)
// }

// func TestNew(t *testing.T) {
// 	mp := MemPool[object]{}
// 	t.Logf("head: %d", mp.cache.header)
// 	pt := &PoolTester{
// 		pool:     &mp,
// 		size:     1 << 16,
// 		batch:    1000,
// 		parallel: 2,
// 	}
// 	pt.pool.Init(pt.size)
// 	ptr := pt.pool.New()
// 	if ptr == nil {
// 		t.Errorf("new -> nil")
// 	}

// 	if pt.pool.Free(ptr) {
// 		t.Logf("Free ok")
// 	} else {
// 		t.Logf("idx: %d", mp.cache.getIndex(ptr))
// 		t.Logf("Free failed")
// 	}
// }

func BenchmarkQueue(b *testing.B) {
	parallel := 2
	runtime.GOMAXPROCS(parallel)
	q := AQueue{}
	// q := chQueue{}
	var size int64 = 1 << 16
	q.Init(size)
	for i := int64(0); i < size; i++ {
		q.Push(i)
	}
	batch := int64(1 << 12)
	data := make([]int64, batch)
	for i := 0; i < int(batch); i++ {
		data[i] = -1
	}

	ch := make(chan int, parallel)

	b.RunParallel(func(pb *testing.PB) {
		failed := 0
		for pb.Next() {
			idx := rand.Int63n(batch)
			if data[idx] >= 0 {
				if q.Push(data[idx]) {
					data[idx] = -1
				} else {

					failed++
				}
			} else {
				value, ok := q.Pop()
				if ok {
					data[idx] = value
				} else {
					failed++
				}
			}
		}
		ch <- failed
	})
	failed := 0
	for i := 0; i < parallel; i++ {
		failed += <-ch
	}

	b.Logf("failed rate: %d / %d = %f", failed, b.N, float64(failed)/float64(b.N))
}

func BenchmarkPool(b *testing.B) {
	parallel := 8
	runtime.GOMAXPROCS(parallel)
	// q := aQueue{}
	// q := isolated.BPointer{}
	q := chQueue{}
	var size int64 = 1 << 16
	q.Init(size)
	for i := int64(0); i < size; i++ {
		q.Push(i)
	}

	batch := int64(1 << 12)

	ch := make(chan [2]int, parallel)

	b.RunParallel(func(pb *testing.PB) {
		data := make([]int64, batch)
		for i := 0; i < int(batch); i++ {
			data[i] = -1
		}
		pushFailed := 0
		popFailed := 0
		var idx int64 = 0
		for pb.Next() {
			idx = (idx + 1) % batch
			if data[idx] >= 0 {
				if q.Push(data[idx]) {
					data[idx] = -1
				} else {
					pushFailed++
				}
			} else {
				value, ok := q.Pop()
				if ok {
					data[idx] = value
				} else {
					popFailed++
				}
			}
		}
		ch <- [2]int{pushFailed, popFailed}
	})
	pushFailed := 0
	popFailed := 0
	for i := 0; i < parallel; i++ {
		r := <-ch
		pushFailed += r[0]
		popFailed += r[1]
	}

	b.Logf("push failed rate: %d / %d = %f", pushFailed, b.N, float64(pushFailed)/float64(b.N))
	b.Logf("pop failed rate: %d / %d = %f", popFailed, b.N, float64(popFailed)/float64(b.N))
}

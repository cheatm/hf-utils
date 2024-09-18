package mempool

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func getParallel(d int) int {
	p, ok := os.LookupEnv("PARALLEL")
	if ok {
		benchParallel, err := strconv.ParseInt(p, 10, 64)
		if err == nil {
			return int(benchParallel)
		}
	}
	return d
}

type iMemPool[T any] interface {
	Free(ptr *T) bool
	New() *T
	Init(int64)
}

type RawMemPool[T any] struct{}

func (p *RawMemPool[T]) Free(*T) bool {
	return true
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

type BenchStats struct {
	AllocCount, AllocFailed, FreeCount, FreeFailed int
}

func (pt *PoolTester) BenchmarkRandomRW(b *testing.B) {
	var count int
	if pt.parallel > 0 {
		runtime.GOMAXPROCS(pt.parallel)
		count = pt.parallel
	} else {
		count = runtime.GOMAXPROCS(0)
	}
	ch := make(chan BenchStats, count)
	pt.pool.Init(pt.size)
	b.RunParallel(func(pb *testing.PB) {
		stats := pt.BenchmarkParallel(pb)
		ch <- stats
	})
	// results := make([]int64, count)
	var totalAlloc, failedAlloc, totalFree, failedFree int
	for i := 0; i < count; i++ {
		// results[i] = <-ch
		stats := <-ch
		totalAlloc = totalAlloc + stats.AllocCount
		failedAlloc = failedAlloc + stats.AllocFailed
		totalFree = totalFree + stats.FreeCount
		failedFree = failedFree + stats.FreeFailed
	}
	b.Logf(
		"N=%d, AllocRate=%f, AllocFailed=%f, FreeFailed=%f",
		b.N, float64(totalAlloc)/float64(b.N),
		float64(failedAlloc)/float64(totalAlloc),
		float64(failedFree)/float64(totalFree),
	)
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	b.Logf("GC Paused: %s", time.Duration(stats.PauseTotalNs))
}

func (pt *PoolTester) BenchmarkParallel(b *testing.PB) BenchStats {
	// pid := pt.id.Add(1)
	array := make([]*object, pt.batch)
	// var allocCount, allocFailed, freeCount, freeFailed int
	var stats BenchStats
	var i int64 = 0
	for b.Next() {
		i++
		_i := i % pt.batch
		if array[_i] == nil {
			stats.AllocCount++
			ptr := pt.pool.New()
			if ptr == nil {
				stats.AllocFailed++
				runtime.Gosched()
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
		} else {
			stats.FreeCount++
			// array[_i].Count = 0
			// array[_i].By = 0
			release := array[_i].Release()
			if release != 0 {
				panic(fmt.Errorf("Release Failed: {i=%d, _i=%d, r=%d}", i, _i, release))
			}
			if pt.pool.Free(array[_i]) {
				array[_i] = nil
			} else {
				stats.FreeFailed++
				array[_i].Require()
				runtime.Gosched()

			}
		}

	}
	return stats
}

func BenchmarkMemPoolRW(b *testing.B) {
	pt := &PoolTester{
		pool:     &MemPool[object]{},
		size:     (1 << 16) - 1,
		batch:    1 << 12,
		parallel: getParallel(4),
	}
	pt.BenchmarkRandomRW(b)
}

func BenchmarkChMemPoolRW(b *testing.B) {
	pt := &PoolTester{
		pool:     &ChMemPool[object]{},
		size:     1 << 16,
		batch:    1 << 12,
		parallel: getParallel(4),
	}
	pt.BenchmarkRandomRW(b)
}

func BenchmarkRawPoolRW(b *testing.B) {
	pt := &PoolTester{
		pool:     &RawMemPool[object]{},
		size:     1 << 16,
		batch:    1 << 12,
		parallel: getParallel(2),
	}
	pt.BenchmarkRandomRW(b)
}

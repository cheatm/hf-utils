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
	return getEnvInt("PARALLEL", d)
}

func getCPU(d int) int {
	return getEnvInt("CPU", d)
}

func getEnvInt(variable string, _default int) int {
	p, ok := os.LookupEnv(variable)
	if ok {
		benchParallel, err := strconv.ParseInt(p, 10, 64)
		if err == nil {
			return int(benchParallel)
		}
	}
	return _default
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

type iObject interface {
	Index() int64
	SetIndex(int64)
	Require() int64
	Release() int64
}

type baseObject[D any] struct {
	Idx   int64
	Data  D
	Count int64
}

func (o *baseObject[D]) Require() int64 {
	return atomic.AddInt64(&o.Count, 1)
}

func (o *baseObject[D]) Release() int64 {
	return atomic.AddInt64(&o.Count, -1)
}

func (o *baseObject[D]) Index() int64 {
	return o.Idx
}

func (o *baseObject[D]) SetIndex(idx int64) {
	o.Idx = idx
}

type object16 = baseObject[[1 << 4]byte]
type object256 = baseObject[[1 << 8]byte]
type object4096 = baseObject[[1 << 12]byte]

type object = object256

type PoolTester[O any, PO interface {
	*O
	iObject
}] struct {
	pool     iMemPool[O]
	size     int64
	batch    int64
	parallel int
	cpus     int
	debug    bool
}

type BenchStats struct {
	AllocCount, AllocFailed, FreeCount, FreeFailed int
	Used                                           int64
}

func (pt *PoolTester[O, PO]) BenchmarkRandomRW(b *testing.B) {
	cpus := pt.cpus
	if pt.cpus > 0 {
		runtime.GOMAXPROCS(cpus)
	} else {
		runtime.GOMAXPROCS(1)
	}
	procs := runtime.GOMAXPROCS(0)
	var count int
	if pt.parallel > 1 {
		count = pt.parallel * procs
		b.SetParallelism(pt.parallel)
	} else {
		count = procs
	}

	ch := make(chan BenchStats, count)
	pt.pool.Init(pt.size)
	b.RunParallel(func(pb *testing.PB) {
		stats := pt.BenchmarkParallel(pb)
		ch <- stats
	})
	// results := make([]int64, count)
	var totalAlloc, failedAlloc, totalFree, failedFree int
	var used int64
	for i := 0; i < count; i++ {
		// results[i] = <-ch
		stats := <-ch
		totalAlloc = totalAlloc + stats.AllocCount
		failedAlloc = failedAlloc + stats.AllocFailed
		totalFree = totalFree + stats.FreeCount
		failedFree = failedFree + stats.FreeFailed
		used += stats.Used
	}
	if pt.debug {
		b.Logf("Procs=%d, Parallel=%d, Count=%d", procs, pt.parallel, count)
		b.Logf(
			"N=%d, U=%d, AllocRate=%f, AllocFailed=%f, FreeFailed=%f",
			b.N, used,
			float64(totalAlloc)/float64(b.N),
			float64(failedAlloc)/float64(totalAlloc),
			float64(failedFree)/float64(totalFree),
		)
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		b.Logf("GC Paused: %s", time.Duration(stats.PauseTotalNs))
	}
}

func (pt *PoolTester[O, PO]) BenchmarkParallel(b *testing.PB) BenchStats {
	// pid := pt.id.Add(1)
	array := make([]PO, pt.batch)
	// var allocCount, allocFailed, freeCount, freeFailed int
	var stats BenchStats
	var i int64 = -1
	for b.Next() {
		i++
		_i := i % pt.batch
		if array[_i] == nil {
			stats.AllocCount++
			ptr := PO(pt.pool.New())
			if ptr == nil {
				stats.AllocFailed++
				runtime.Gosched()
				continue
			}
			ptr.SetIndex(_i)
			required := ptr.Require()
			if required != 1 {
				panic(fmt.Errorf("Require Failed: {i=%d, _i=%d, r=%d}", i, _i, required))
			}
			array[_i] = ptr
		} else {
			stats.FreeCount++
			ptr := array[_i]
			release := ptr.Release()
			if release != 0 {
				panic(fmt.Errorf("Release Failed: {i=%d, _i=%d, r=%d}", i, _i, release))
			}
			tmp := ptr.Index()
			stats.Used += tmp
			ptr.SetIndex(0)
			if pt.pool.Free((*O)(array[_i])) {
				array[_i] = nil
			} else {
				stats.FreeFailed++
				array[_i].Require()
				ptr.SetIndex(tmp)
				stats.Used -= tmp
				runtime.Gosched()

			}
		}
	}

	return stats
}

type MultiTester[O any, PO interface {
	*O
	iObject
}] struct {
	name   string
	makers map[string]func() iMemPool[O]
	cp     [][2]int // cpu & parallel
	size   int64
	batch  int64
}

func (m *MultiTester[O, PO]) Benchmark(b *testing.B) {

	for poolType, maker := range m.makers {
		for _, cp := range m.cp {
			pt := &PoolTester[O, PO]{
				pool:     maker(),
				size:     m.size,
				batch:    m.batch,
				cpus:     cp[0],
				parallel: cp[1],
			}
			b.Run(fmt.Sprintf("%s-%s-%dC-%dP", m.name, poolType, pt.cpus, pt.parallel), pt.BenchmarkRandomRW)
		}
	}
}

func newMemPool[O any]() iMemPool[O] {
	return new(MemPool[O])
}

func newChPool[O any]() iMemPool[O] {
	return new(ChMemPool[O])
}

func newRawPool[O any]() iMemPool[O] {
	return new(RawMemPool[O])
}

func BenchmarkMemPoolRW(b *testing.B) {
	pt := &PoolTester[object, *object]{
		pool:     &MemPool[object]{},
		size:     (1 << 16),
		batch:    1 << 12,
		parallel: getParallel(1),
		cpus:     getCPU(2),
		debug:    false,
	}
	pt.BenchmarkRandomRW(b)
}

func BenchmarkChMemPoolRW(b *testing.B) {
	pt := &PoolTester[object, *object]{
		pool:     &ChMemPool[object]{},
		size:     1 << 16,
		batch:    1 << 12,
		parallel: getParallel(4),
		cpus:     getCPU(1),
		debug:    true,
	}
	pt.BenchmarkRandomRW(b)
}

func BenchmarkRawPoolRW(b *testing.B) {
	pt := &PoolTester[object, *object]{
		pool:     &RawMemPool[object]{},
		size:     1 << 16,
		batch:    1 << 12,
		parallel: getParallel(2),
		cpus:     getCPU(2),
		debug:    true,
	}
	pt.BenchmarkRandomRW(b)
}

var benchmarkCPs = [][2]int{
	{1, 1},
	{2, 1},
	{3, 1},
	{4, 1},
	{5, 1},
	{6, 1},
}

func BenchmarkMultiObj16RW(b *testing.B) {
	pt := &MultiTester[object16, *object16]{
		name: "obj-16B",
		makers: map[string]func() iMemPool[object16]{
			"casq": newMemPool[object16],
			"chan": newChPool[object16],
			"raw":  newRawPool[object16],
		},
		size:  (1 << 18),
		batch: 1 << 12,
		cp:    benchmarkCPs,
	}
	pt.Benchmark(b)
}

func BenchmarkMultiObj256RW(b *testing.B) {
	pt := &MultiTester[object256, *object256]{
		name: "obj-256B",
		makers: map[string]func() iMemPool[object256]{
			"casq": newMemPool[object256],
			"chan": newChPool[object256],
			"raw":  newRawPool[object256],
		},
		size:  (1 << 18),
		batch: 1 << 12,
		cp:    benchmarkCPs,
	}
	pt.Benchmark(b)
}

func BenchmarkMultiObj4096RW(b *testing.B) {
	pt := &MultiTester[object4096, *object4096]{
		name: "obj-4096B",
		makers: map[string]func() iMemPool[object4096]{
			"casq": newMemPool[object4096],
			"chan": newChPool[object4096],
			"raw":  newRawPool[object4096],
		},
		size:  (1 << 18),
		batch: 1 << 12,
		cp:    benchmarkCPs,
	}
	pt.Benchmark(b)
}

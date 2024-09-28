package mempool

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

type ShardTester[O any, PO interface {
	*O
	iObject
}] struct {
	maker    func() iMemPool[O]
	pools    []iMemPool[O]
	shards   int64
	shSize   int64
	batch    int64
	parallel int
	cpus     int
	debug    bool
}

func (s *ShardTester[O, PO]) BenchmarkParallel(b *testing.B) {
	cpus := s.cpus
	if s.cpus > 0 {
		runtime.GOMAXPROCS(cpus)
	} else {
		runtime.GOMAXPROCS(1)
	}
	procs := runtime.GOMAXPROCS(0)
	var count int
	if s.parallel > 1 {
		count = s.parallel * procs
		b.SetParallelism(s.parallel)
	} else {
		count = procs
	}

	ch := make(chan BenchStats, count)
	// pt.pool.Init(pt.size)
	s.pools = make([]iMemPool[O], s.shards)
	for i := 0; i < int(s.shards); i++ {
		s.pools[i] = s.maker()
		s.pools[i].Init(s.shSize)
	}

	b.RunParallel(func(pb *testing.PB) {
		stats := s.Parallel(pb)
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
	if s.debug {
		b.Logf("Procs=%d, Parallel=%d, Count=%d", procs, s.parallel, count)
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

func (s *ShardTester[O, PO]) Parallel(b *testing.PB) BenchStats {
	array := make([]PO, s.batch)
	// var allocCount, allocFailed, freeCount, freeFailed int
	var stats BenchStats
	var i int64 = -1
	for b.Next() {
		i++
		_i := i % s.batch
		shard := i % s.shards
		if array[_i] == nil {
			stats.AllocCount++
			ptr := PO(s.pools[shard].New())
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
			if s.pools[shard].Free((*O)(array[_i])) {
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

func BenchmarkShardCasQ(b *testing.B) {
	tester := ShardTester[object256, *object256]{
		maker:    newMemPool[object],
		shards:   2,
		shSize:   1 << 13,
		batch:    1 << 12,
		parallel: getParallel(1),
		cpus:     getCPU(4),
		debug:    true,
	}

	tester.BenchmarkParallel(b)
}

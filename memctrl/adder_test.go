package memctrl

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

type IQueue interface {
	Push(int64)
	Pop() int64
}

type qTester struct {
	q     IQueue
	p     int
	batch int
}

func (q *qTester) BenchmarkParallel(b *testing.B) {
	if q.p > 0 {
		runtime.GOMAXPROCS(q.p)
	}
	ch := make(chan int64, 256)
	var count int64
	b.RunParallel(func(pb *testing.PB) {
		atomic.AddInt64(&count, 1)
		result := q.ParallelLoop(pb)
		ch <- result
	})
	results := make([]int64, int(count))
	for i := 0; i < int(count); i++ {
		results[i] = <-ch
	}

	b.Logf("results: %v", results)
}

type PB struct {
	N int64
	n int64
}

func (pb *PB) Next() bool {
	return atomic.AddInt64(&pb.n, 1) <= pb.N
}

func (q *qTester) BenchmarkGo(b *testing.B) {
	results := make([]int64, int(q.p))
	wg := sync.WaitGroup{}
	pb := PB{N: int64(b.N), n: 0}
	for i := 0; i < q.p; i++ {
		wg.Add(1)
		go func(c int) {
			results[c] = q.ParallelLoop(&pb)
			wg.Done()
		}(i)
	}

	wg.Wait()
	b.Logf("results: %v", results)
}

func (q *qTester) ParallelLoop(b interface{ Next() bool }) int64 {
	var (
		temp   int64
		result int64
		_b     int
	)
	w := true
	var i int64 = 0
	for b.Next() {
		if w {
			q.q.Push(i)
			i++
		} else {
			newTemp := q.q.Pop()
			result += newTemp - temp
			temp = newTemp
		}
		_b++
		if _b >= q.batch {
			w = !w
		}
	}
	return result
}

func BenchmarkPointer(b *testing.B) {
	var q IQueue = &Pointer{size: 1024}
	qt := &qTester{q: q, p: 1, batch: 64}
	qt.BenchmarkParallel(b)
	b.Logf("w: %d, r: %d, f: %d", q.(*Pointer).w, q.(*Pointer).r, q.(*Pointer).failed)
}

func BenchmarkPointerGo(b *testing.B) {
	var q IQueue = &Pointer{size: 1024}
	qt := &qTester{q: q, p: 2, batch: 64}
	qt.BenchmarkParallel(b)
	b.Logf("w: %d, r: %d, f: %d", q.(*Pointer).w, q.(*Pointer).r, q.(*Pointer).failed)
}

func BenchmarkChan(b *testing.B) {
	var q IQueue = &ChCycle{ch: make(chan int64, 1024)}
	qt := &qTester{q: q, p: 1, batch: 64}
	qt.BenchmarkParallel(b)
}

func BenchmarkAdder(b *testing.B) {
	// var q IQueue = &Pointer{size: 1024}
	var q IQueue = &Adder{s: 1024}
	qt := &qTester{q: q, p: 2, batch: 64}
	qt.BenchmarkParallel(b)
	b.Logf("w: %d, r: %d, f: %d", q.(*Adder).w, q.(*Adder).r, q.(*Adder).f)
}

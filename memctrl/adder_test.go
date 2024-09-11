package memctrl

import (
	"runtime"
	"sync/atomic"
	"testing"
)

type IQueue interface {
	Push()
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

func (q *qTester) ParallelLoop(b *testing.PB) int64 {
	var (
		temp   int64
		result int64
		_b     int
	)
	w := true
	for b.Next() {
		if w {
			q.q.Push()
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
	qt := &qTester{q: q, p: 2, batch: 64}
	qt.BenchmarkParallel(b)
	b.Logf("w: %d, r: %d, f: %d", q.(*Pointer).w, q.(*Pointer).r, q.(*Pointer).failed)
}

func BenchmarkAdder(b *testing.B) {
	// var q IQueue = &Pointer{size: 1024}
	var q IQueue = &Adder{s: 1024}
	qt := &qTester{q: q, p: 2, batch: 64}
	qt.BenchmarkParallel(b)
	b.Logf("w: %d, r: %d, f: %d", q.(*Adder).w, q.(*Adder).r, q.(*Adder).f)
}

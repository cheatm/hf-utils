package mempool

import (
	"fmt"
	"runtime"
	"testing"
)

type iQueue interface {
	Init(int64)
	Push(int64) bool
	Pop() (int64, bool)
}

type QueueTester struct {
	queue    iQueue
	parallel int
	batch    int64
	size     int64
}

type BenchResult struct {
	PushCount  int
	PopCount   int
	PushFailed int
	PopFailed  int
}

func (qt *QueueTester) BenchmarkPool(b *testing.B) {
	parallel := qt.parallel
	runtime.GOMAXPROCS(parallel)
	q := qt.queue
	// var size int64 = 1 << 16
	size := qt.size
	q.Init(size)
	for i := int64(0); i < size; i++ {
		q.Push(i)
	}
	batch := qt.batch

	ch := make(chan BenchResult, parallel)

	b.RunParallel(func(pb *testing.PB) {
		data := make([]int64, batch)
		for i := 0; i < int(batch); i++ {
			data[i] = -1
		}
		br := BenchResult{}
		var idx int64 = 0
		for pb.Next() {
			idx = (idx + 1) % batch
			if data[idx] >= 0 {
				br.PushCount++
				if q.Push(data[idx]) {
					data[idx] = -1
				} else {
					br.PushFailed++
				}
			} else {
				br.PopCount++
				value, ok := q.Pop()
				if ok {
					data[idx] = value
				} else {
					br.PopFailed++
				}
			}
		}
		ch <- br
	})
	br := BenchResult{}
	for i := 0; i < parallel; i++ {
		tmp := <-ch
		br.PopCount += tmp.PopCount
		br.PushCount += tmp.PushCount
		br.PopFailed += tmp.PopFailed
		br.PushFailed += tmp.PushFailed
	}

	b.Logf("push failed rate: %d / %d = %f", br.PushFailed, br.PushCount, float64(br.PushFailed)/float64(br.PushCount))
	b.Logf("pop failed rate: %d / %d = %f", br.PopFailed, br.PopCount, float64(br.PopFailed)/float64(br.PopCount))
}

func BenchmarkAQueue(b *testing.B) {
	qt := QueueTester{
		queue:    &aQueue{},
		parallel: 2,
		size:     1 << 16,
		batch:    1 << 12,
	}

	qt.BenchmarkPool(b)
}

func BenchmarkCQueue(b *testing.B) {
	qt := QueueTester{
		queue:    &chQueue{},
		parallel: 2,
		size:     1 << 16,
		batch:    1 << 12,
	}

	qt.BenchmarkPool(b)
}

type ParallelQueueTester struct {
	maker     func() iQueue
	parallels []int
	batch     int64
	size      int64
}

func (pt *ParallelQueueTester) BenchmarkPool(b *testing.B) {
	for _, parallel := range pt.parallels {
		qt := QueueTester{
			queue:    pt.maker(),
			size:     pt.size,
			batch:    pt.batch,
			parallel: parallel,
		}
		b.Run(fmt.Sprintf("Parallel-%d", parallel), qt.BenchmarkPool)
	}
}

func BenchmarkAQueueParallels(b *testing.B) {
	pqt := ParallelQueueTester{
		maker:     func() iQueue { return &aQueue{} },
		parallels: []int{1, 2, 4, 8},
		batch:     1 << 12,
		size:      1 << 16,
	}
	pqt.BenchmarkPool(b)
}

func BenchmarkCQueueParallels(b *testing.B) {
	pqt := ParallelQueueTester{
		maker:     func() iQueue { return &chQueue{} },
		parallels: []int{1, 2, 4, 8},
		batch:     1 << 12,
		size:      1 << 16,
	}
	pqt.BenchmarkPool(b)
}

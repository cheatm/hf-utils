package mempool

import (
	"runtime"
	"sync/atomic"
	"testing"
)

type Cache struct {
	Data [1024]int64
}

func (c *Cache) Add(pos int) {
	atomic.AddInt64(&c.Data[pos], 1)
}

func BenchmarkCacheLine(b *testing.B) {
	indexes := []int{0b0111, 0b1000}
	proc := len(indexes)
	runtime.GOMAXPROCS(proc)
	c := Cache{}
	idx := atomic.Int64{}
	b.RunParallel(func(p *testing.PB) {
		pos := indexes[idx.Add(1)-1]
		for p.Next() {
			c.Add(pos)
		}
	})
	for _, i := range indexes {

		b.Logf("%d: %d", i, c.Data[i])
	}
}

package mempool

import (
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"
)

type ACache struct {
	Data [1024]int64
}

func (c *ACache) Add(pos int) {
	atomic.AddInt64(&c.Data[pos], 1)
}

type SCache struct {
	Data [1024]int64
}

func (c *SCache) Add(pos int) {
	c.Data[pos] += 1
}

func BenchmarkCacheLine(b *testing.B) {
	indexes := []int{getEnvInt("C0", 0), getEnvInt("C1", 1)}
	proc := len(indexes)
	runtime.GOMAXPROCS(proc)
	// c := ACache{}
	c := SCache{}
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

func getSliceHead[T any](array []T) uintptr {
	return (*reflect.SliceHeader)(unsafe.Pointer(&array)).Data
}

func TestSliceHeader(t *testing.T) {
	array := make([]int64, 0, 2)
	t.Logf("head: %d", getSliceHead(array))
	array = append(array, 1)
	t.Logf("head: %d", getSliceHead(array))
	array = append(array, 2)
	t.Logf("head: %d", getSliceHead(array))
	array = append(array, 3)
	t.Logf("head: %d", getSliceHead(array))
	t.Logf("cap: %d", cap(array))
}

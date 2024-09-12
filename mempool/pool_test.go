package mempool

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

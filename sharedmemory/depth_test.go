package sharedmemory

import (
	"fmt"
	"math/rand"
	"testing"

	"bitbucket.org/avd/go-ipc/shm"
	"github.com/stretchr/testify/assert"
)

const DEPTH_MEM_NAME string = "DepthM"
const DEPTH_MEM_TRUNCATE int64 = 1 << 8

var dm shm.SharedMemoryObject

func initDepthMem() {
	if dm == nil {
		dm = initMemoryObject(DEPTH_MEM_NAME)
	}
}

func createDepthMem() {
	if dm == nil {
		dm = createMemoryObject(DEPTH_MEM_NAME, DEPTH_MEM_TRUNCATE*(int64(FullDepthSize())))
	}
}

func TestDepthSize(t *testing.T) {
	t.Logf("FullDepth size: %d", FullDepthSize())
}

func TestDepthSliceSize(t *testing.T) {
	depths := make([]FullDepth, 2)
	depths[0].Price = 1
	depths[0].Asks[0].Volume = 2
	elemSize := FullDepthSize()
	dBytes := UnsafeSliceToBytes(depths, elemSize)
	t.Logf("Price: %d", depths[0].Price)
	t.Logf("dBytes: %v", dBytes[:100])

}

func TestWR(t *testing.T) {
	createDepthMem()
	defer shm.DestroyMemoryObject(DEPTH_MEM_NAME)
	table := NewSharedMemoryTable[FullDepth](dm)
	table.Chunk = 1 << 4
	depths := make([]FullDepth, 10)
	depths[0].Price = 1
	depths[0].Asks[0].Volume = 2
	depths[1].Price = 3

	loc, err := table.Write(0, depths)
	if err != nil {
		t.Fatalf("Err: %s", err)
	}
	t.Logf("Write loc at: %d", loc)

	newDepths, err := table.Read(0, 10)
	if err != nil {
		t.Fatalf("Err: %s", err)
	}
	assert.Equal(t, newDepths[0].Price, depths[0].Price)
	assert.Equal(t, newDepths[0].Asks[0].Volume, depths[0].Asks[0].Volume)
	assert.Equal(t, newDepths[1].Price, depths[1].Price)

}

func SetRandomDepth(depth *FullDepth, price int64, timestamp int64) {
	askVolume := rand.Int63n(300)
	bidVolume := rand.Int63n(300)
	depth.Price = price
	depth.Timestamp = timestamp

	for i := 0; i < len(depth.Asks); i++ {
		depth.Asks[i].Price = price + int64(i) + 1
		depth.Asks[i].Volume = askVolume
		askVolume--
	}

	for i := 0; i < len(depth.Bids); i++ {
		depth.Bids[i].Price = price - int64(i)
		depth.Bids[i].Volume = bidVolume
		bidVolume--
	}

}

type DepthGenerator struct {
	Price     int64
	Range     int64
	Offset    int64
	Timestamp int64
}

func (dg *DepthGenerator) Next(depth *FullDepth) {
	SetRandomDepth(depth, dg.Price, dg.Timestamp)
	dg.Timestamp++
	tick := rand.Int63n(dg.Range) - dg.Offset
	dg.Price = dg.Price + tick
}

func TestRandomDepth(t *testing.T) {
	dg := DepthGenerator{
		Price:  10000,
		Range:  11,
		Offset: 5,
	}
	depths := make([]FullDepth, 10)
	for i := 0; i < len(depths); i++ {
		dg.Next(&(depths[i]))
	}
}

func BenchmarkGenDepth(b *testing.B) {
	size := 128
	dg := DepthGenerator{
		Price:  10000,
		Range:  11,
		Offset: 5,
	}
	depths := make([]FullDepth, size)
	for i := 0; i < b.N; i++ {
		dg.Next(&depths[i%size])
	}
}

func TestDepthWrite(t *testing.T) {
	createDepthMem()
	table := NewSharedMemoryTable[FullDepth](dm)
	size := 64
	table.Chunk = size
	cap := table.smo.Size() / int64(table.elemSize)
	dg := DepthGenerator{
		Price:  10000,
		Range:  11,
		Offset: 5,
	}
	depths := make([]FullDepth, size)
	for offset := 0; offset+size < int(cap); offset = offset + size {
		for i := 0; i < size; i++ {
			dg.Next(&depths[i])
		}
		_, err := table.Write(offset, depths)
		if err != nil {
			panic(err)
		}
	}
}

func TestDepthRead(t *testing.T) {
	initDepthMem()
	table := NewSharedMemoryTable[FullDepth](dm)
	size := 64
	data, err := table.Read(0, size)
	if err != nil {
		panic(err)
	}
	t.Logf("Data[0].Price = %d", data[0].Price)
	t.Logf("Data[0].Times = %d", data[0].Timestamp)
	t.Logf("Data[0].Asks[0] = [%d:%d]", data[0].Asks[0].Price, data[0].Asks[0].Volume)
	t.Logf("Data[0].Bids[0] = [%d:%d]", data[0].Bids[0].Price, data[0].Bids[0].Volume)

}

type DepthTestParam struct {
	Size int
}

func (p *DepthTestParam) BenchmarkRead(b *testing.B) {
	initDepthMem()
	defer func() {
		dm.Close()
		dm = nil
	}()

	table := NewSharedMemoryTable[FullDepth](dm)
	table.Chunk = p.Size
	offset := 0
	cap := table.Cap()
	for i := 0; i < b.N; i++ {
		if i%p.Size == 0 {
			if offset+p.Size > cap {
				offset = 0
			}
			table.Read(offset, p.Size)
			offset = offset + p.Size
		}
	}
}

func BenchmarkDepthRead(b *testing.B) {
	for _, chunk := range []int{32, 64, 128} {
		b.Run(
			fmt.Sprintf("Chunk=%d", chunk),
			(&DepthTestParam{chunk}).BenchmarkRead,
		)
	}
}

func BenchmarkDepthWrite(b *testing.B) {
	createDepthMem()
	defer shm.DestroyMemoryObject(DEPTH_MEM_NAME)
	table := NewSharedMemoryTable[FullDepth](dm)
	size := 64
	table.Chunk = size
	dg := DepthGenerator{
		Price:  10000,
		Range:  11,
		Offset: 5,
	}
	depths := make([]FullDepth, size)
	wCount := 0
	for i := 0; i < b.N; i++ {
		if i != 0 && i%size == 0 {
			_, err := table.Write(0, depths)
			if err != nil {
				panic(err)
			}
			wCount++
		}
		dg.Next(&depths[i%size])

	}
	b.Logf("N: %d, W: %d", b.N, wCount)

}

package sharedmemory

import (
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

package sharedmemory

import (
	"fmt"
	"reflect"
	"unsafe"

	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
)

type UnsafeSharedMemoryTable[T any] struct {
	elemSize int
	smo      shm.SharedMemoryObject
	Chunk    int
}

func (table *UnsafeSharedMemoryTable[T]) Close() error {
	return table.smo.Close()
}

func NewSharedMemoryTable[T any](obj shm.SharedMemoryObject) *UnsafeSharedMemoryTable[T] {
	return &UnsafeSharedMemoryTable[T]{
		elemSize: int(reflect.TypeOf(new(T)).Elem().Size()),
		smo:      obj,
		Chunk:    128,
	}
}

func (table *UnsafeSharedMemoryTable[T]) Read(offset int, size int) (data []T, err error) {
	locEnd := (int64(offset) + int64(size)) * int64(table.elemSize)
	if locEnd > table.smo.Size() {
		err = fmt.Errorf("size execeed memory total size")
		return
	}
	var roRegion *mmf.MemoryRegion
	loc := int64(offset) * int64(table.elemSize)
	tBytes := make([]byte, size*table.elemSize)
	pos := 0
	for pos < size {
		endPos := pos + table.Chunk
		if endPos > size {
			endPos = size
		}
		chunkSize := (endPos - pos) * table.elemSize
		roRegion, err = mmf.NewMemoryRegion(table.smo, mmf.MEM_READ_ONLY, loc, chunkSize)
		if err != nil {
			return
		}
		copy(tBytes[pos*table.elemSize:endPos*table.elemSize], roRegion.Data())
		roRegion.Close()
		pos = endPos
		loc = loc + int64(chunkSize)
	}

	data = UnsafeBytesToSlice[T](tBytes, table.elemSize)
	return
}

func (table *UnsafeSharedMemoryTable[T]) Write(offset int, data []T) (loc int64, err error) {
	length := len(data)
	totalSize := table.smo.Size()
	loc = int64(offset) * int64(table.elemSize)
	if loc+int64(length)*int64(table.elemSize) > totalSize {
		err = fmt.Errorf("size execeed memory total size")
		return
	}
	var rwRegion *mmf.MemoryRegion
	var wCount int
	pos := 0
	for pos < length {
		posEnd := pos + table.Chunk
		if posEnd > length {
			posEnd = length
		}
		chunk := data[pos:posEnd]
		chunkBytes := UnsafeSliceToBytes(chunk, table.elemSize)
		chunkSize := len(chunkBytes)
		rwRegion, err = mmf.NewMemoryRegion(table.smo, mmf.MEM_READWRITE, loc, chunkSize)
		if err != nil {
			return
		}
		wCount, err = mmf.NewMemoryRegionWriter(rwRegion).Write(chunkBytes)
		if err != nil {
			return
		}
		loc = loc + int64(wCount)

		rwRegion.Close()
		pos = posEnd
	}

	return
}

func UnsafeSliceToBytes[T any](data []T, elemSize int) []byte {
	up := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	var result []byte

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&result))
	sh.Cap = elemSize * len(data)
	sh.Len = sh.Cap
	sh.Data = up.Data
	return result
}

func UnsafeBytesToSlice[T any](data []byte, elemSize int) []T {
	up := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	var result []T
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&result))
	sh.Cap = len(data) / elemSize
	sh.Len = sh.Cap
	sh.Data = up.Data
	return result
}

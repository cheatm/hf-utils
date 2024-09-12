package mempool

import (
	"reflect"
	"sync/atomic"
	"unsafe"
)

type bitmapCache[T any] struct {
	cache    []T
	tag      []atomic.Bool
	size     int64
	header   uintptr
	elemSize uintptr
}

func (c *bitmapCache[T]) getIndex(t *T) uintptr {
	ptr := uintptr(unsafe.Pointer(t))
	return (ptr - c.header) / c.elemSize
}

func (c *bitmapCache[T]) init(size int64) {
	c.cache = make([]T, size)
	c.tag = make([]atomic.Bool, size)
	c.size = size
	c.header = uintptr(unsafe.Pointer(&c.cache[0]))
	var t T
	c.elemSize = reflect.TypeOf(t).Size()
}

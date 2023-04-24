package sharedmemory

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestInt64ToBytes(t *testing.T) {
	i := int64(0x11F1)
	t.Logf("int: %d", i)
	usptr := unsafe.Pointer(&i)
	bytes := (*[8]byte)(usptr)

	for _, b := range *bytes {
		t.Logf("%d", b)
	}
	iptr := (*int64)(unsafe.Pointer(bytes))

	t.Logf("int: %d", *iptr)

}

type TwoInt struct {
	i0 int64
	i1 int64
}

func Test2Int(t *testing.T) {
	ti := &TwoInt{4, 10}
	bytes := (*[16]byte)(unsafe.Pointer(ti))
	t.Logf("barray: %v", bytes)
	ti = (*TwoInt)(unsafe.Pointer(bytes))
	t.Logf("twoint: %v", ti)
}

func Test2IntMake(t *testing.T) {
	ti := &TwoInt{4, 10}
	up := uintptr(unsafe.Pointer(ti))

	var data []byte

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	sh.Cap = 16
	sh.Len = 16
	sh.Data = up
	t.Logf("Data: %v", data)
}

func Test2IntMakeSlice(t *testing.T) {
	tis := []TwoInt{TwoInt{4, 10}, TwoInt{5, 3}}
	up := uintptr(unsafe.Pointer(&tis))

	var data []byte

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	sh.Cap = 16 * len(tis)
	sh.Len = 16 * len(tis)
	sh.Data = up
	t.Logf("Data: %v", data)
}

func TestSize(t *testing.T) {
	s := "ab"
	v := reflect.ValueOf(s)
	size := int(v.Type().Size()) * v.Len()
	bytes := (*[]byte)(unsafe.Pointer(&s))
	t.Logf("size: %d", len(*bytes))
	t.Logf("bytes: %v", (*bytes)[:size])
	sptr := (*string)(unsafe.Pointer(bytes))

	t.Logf("string[%d]: %s", len(*sptr), *sptr)
}

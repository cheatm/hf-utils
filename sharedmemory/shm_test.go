package sharedmemory

import (
	"os"
	"runtime"
	"testing"
	"time"

	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

const SHM_NAME string = "shmtest"
const SHM_TRUNCATE int64 = 1 << 23

var mo shm.SharedMemoryObject

func createMemoryObject(name string, truncate int64) shm.SharedMemoryObject {
	shm.DestroyMemoryObject(name)
	obj, err := shm.NewMemoryObject(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		panic(errors.Wrap(err, "create memory object"))
	}
	if err := obj.Truncate(truncate); err != nil {
		panic(errors.Wrap(err, "truncate"))
	}
	return obj
}

func initMemoryObject(name string) shm.SharedMemoryObject {
	obj, err := shm.NewMemoryObject(name, os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		panic(errors.Wrap(err, "init memory object"))
	}
	return obj
}

func createTestMem() {
	if mo == nil {
		mo = createMemoryObject(SHM_NAME, SHM_TRUNCATE)
	}
}

func initTestMem() {
	if mo == nil {
		mo = initMemoryObject(SHM_NAME)
	}
}

func TestSample(t *testing.T) {
	RWSample()
}

func TestObject(t *testing.T) {
	// cleanup previous objects
	shm.DestroyMemoryObject("obj")
	// create new object and resize it.
	obj, err := shm.NewMemoryObject("obj", os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		panic("new")
	}
	defer obj.Close()
	fd := obj.Fd()
	t.Logf("FD: %d\n", fd)
	if err != nil {
		panic("new")
	}
	if err := obj.Truncate(1 << 20); err != nil {
		panic("truncate")
	}
	t.Logf("Size: %d\n", obj.Size())
}

func TestWrite(t *testing.T) {
	obj, err := shm.NewMemoryObject("obj", os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer obj.Close()
	t.Logf("FD: %d\n", obj.Fd())
	t.Logf("Size: %d\n", obj.Size())
	obj.Size()
	// create two regions for reading and writing.
	rwRegion, err := mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, 1024)
	if err != nil {
		panic("new region")
	}
	defer rwRegion.Close()
	writer := mmf.NewMemoryRegionWriter(rwRegion)
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	written, err := writer.WriteAt(data, 0)
	if err != nil || written != len(data) {
		panic("write")
	}
	t.Logf("Write: %d\n", written)
}

func BenchmarkWrite(b *testing.B) {
	obj, err := shm.NewMemoryObject("obj", os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer obj.Close()
	b.Logf("FD: %d\n", obj.Fd())
	b.Logf("Size: %d\n", obj.Size())
	truncsize := obj.Size()
	// create two regions for reading and writing.
	rwRegion, err := mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, int(truncsize))
	if err != nil {
		panic("new region")
	}
	defer rwRegion.Close()
	writer := mmf.NewMemoryRegionWriter(rwRegion)
	var size int64 = 1 << 12
	var offset int64 = 0
	data := make([]byte, size)
	for i := 0; i < b.N; i++ {
		written, err := writer.WriteAt(data, offset)
		if err != nil || written != len(data) {
			panic("write")
		}
		offset = (offset + size) % truncsize
	}
}

func TestRead(t *testing.T) {
	obj, err := shm.NewMemoryObject("obj", os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer obj.Close()
	t.Logf("FD: %d\n", obj.Fd())
	t.Logf("Size: %d\n", obj.Size())

	roRegion, err := mmf.NewMemoryRegion(obj, mmf.MEM_READ_ONLY, 0, 1024)
	if err != nil {
		panic("new region")
	}
	defer roRegion.Close()
	reader := mmf.NewMemoryRegionReader(roRegion)
	actual := make([]byte, 7)
	read, err := reader.ReadAt(actual, 1)
	if err != nil {
		panic(err)
	}
	t.Logf("Read %d\n", read)
	for _, v := range actual {
		t.Logf("%d ", v)
	}
}

func BenchmarkRead(b *testing.B) {
	obj, err := shm.NewMemoryObject("obj", os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer obj.Close()

	truncsize := obj.Size()

	roRegion, err := mmf.NewMemoryRegion(obj, mmf.MEM_READ_ONLY, 0, int(truncsize))
	if err != nil {
		panic("new region")
	}
	defer roRegion.Close()

	var size int64 = 1 << 12
	var offset int64 = 0

	reader := mmf.NewMemoryRegionReader(roRegion)
	actual := make([]byte, size)
	for i := 0; i < b.N; i++ {
		read, err := reader.ReadAt(actual, offset)
		if err != nil || read != len(actual) {
			panic(err)
		}
		offset = (offset + size) % truncsize

	}

}

func TestWriteString(t *testing.T) {
	createTestMem()
	defer mo.Close()

	rwRegion, err := mmf.NewMemoryRegion(mo, mmf.MEM_READWRITE, 0, 2048)
	if err != nil {
		panic("new region")
	}

	defer rwRegion.Close()
	// writer := mmf.NewMemoryRegionWriter(rwRegion)
	// defer rwRegion.Close()
	// rwRegion.

	data := rwRegion.Data()
	t.Logf("Size of region: %d", len(data))
	data[0] = 1
}

func TestReadString(t *testing.T) {
	initTestMem()
	defer mo.Close()

	rdRegion, err := mmf.NewMemoryRegion(mo, mmf.MEM_READ_ONLY, 0, 2048)
	if err != nil {
		panic("new region")
	}

	defer rdRegion.Close()

	data := rdRegion.Data()
	t.Logf("data[1] = %d", data[1])
}

func TestDestroy(t *testing.T) {
	shm.DestroyMemoryObject(SHM_NAME)
}

func TestRuntime(t *testing.T) {
	t.Logf("Runtime: %s", runtime.GOOS)
	t.Logf("uid: %d", unix.Geteuid())
}

const RWNAME = "RWTest"

func TestRW(t *testing.T) {
	memw := createMemoryObject(RWNAME, 4096)
	defer func() {
		memw.Close()
		shm.DestroyMemoryObject(RWNAME)
	}()
	memr := initMemoryObject(RWNAME)
	defer memr.Close()

	rwRegion, err := mmf.NewMemoryRegion(memw, mmf.MEM_READWRITE, 0, 4096)
	if err != nil {
		t.Fatal(err)
	}
	defer rwRegion.Close()

	roRegion, err := mmf.NewMemoryRegion(memr, mmf.MEM_READ_ONLY, 0, 4096)
	if err != nil {
		t.Fatal(err)
	}
	defer roRegion.Close()

	t.Logf("Origin  r[0] = %d", roRegion.Data()[0])
	rwRegion.Data()[0] = 1
	rwRegion.Flush(true)
	t.Logf("Flushed r[0] = %d", roRegion.Data()[0])
	rwRegion.Data()[0] = 2
	t.Logf("Noflush r[0] = %d", roRegion.Data()[0])

}

func TestHold(t *testing.T) {
	memw := createMemoryObject(RWNAME, 4096)
	defer func() {
		memw.Close()
		shm.DestroyMemoryObject(RWNAME)
	}()

	rwRegion, err := mmf.NewMemoryRegion(memw, mmf.MEM_READWRITE, 0, 4096)
	if err != nil {
		t.Fatal(err)
	}

	defer rwRegion.Close()
	expireAt := time.Now().Unix() + 50
	var i uint8 = 0
	for time.Now().Unix() < expireAt {
		time.Sleep(5 * time.Second)
		rwRegion.Data()[0] = i
		t.Logf("i: %d", i)
		i++
	}
}

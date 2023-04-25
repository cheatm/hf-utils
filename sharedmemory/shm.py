from multiprocessing import shared_memory, resource_tracker, managers
# import _posixshmem
import os
import time


def bytes2int64(bts: bytes) -> int :
    r = 0
    for i in range(8):
        r = r + (int(bts[i]) << (i*8))
    return r


def darwin_name(name: str) -> str:
    return "%s\t501" % name


def decode_depth(text: bytes) -> dict: 
    if len(text) < 3216:
        raise ValueError("len(text) < 3216")
    ask_start = 16
    bid_start = 16 + 16*100
    result = {
        "price": int.from_bytes(reversed(text[:8]), "big", signed=True),
        "timestamp": int.from_bytes(reversed(text[8:16]), "big", signed=True),
        "asks": [[
            int.from_bytes(reversed(text[ask_start+i*16:ask_start+8+i*16]), "big", signed=True), 
            int.from_bytes(reversed(text[ask_start+8+i*16:ask_start+16+i*16]), "big", signed=True)
        ] for i in range(100)],
        "bids": [[
            int.from_bytes(reversed(text[bid_start+i*16:bid_start+8+i*16]), "big", signed=True), 
            int.from_bytes(reversed(text[bid_start+8+i*16:bid_start+16+i*16]), "big", signed=True)
        ] for i in range(100)],
    }
    return result


def read_sm(mem: shared_memory.SharedMemory):
    buffer = mem.buf
    start = time.time()
    result = [decode_depth(bytes(buffer[i:i+3216])) for i in range(0, len(buffer), 3216)]
    # result = bytes(buffer[:])
    end = time.time()
    print(len(result))
    print("decode one", (end - start)*1e6/256, "us")


def test_sm(name: str = "DepthM"):
    dname = darwin_name(name)
    print(dname)
    mem = shared_memory.SharedMemory(dname)
    try:
        read_sm(mem)
    finally:
        mem.close()
        resource_tracker.unregister(mem._name, "shared_memory")

def test_manager(name: str="DepthM"):
    dname = darwin_name(name)
    manager  = manager.SharedMemoryManager()
    manager.start()
    mem = manager.SharedMemory()

def test_os():
    print(os.name)
    

def test_read0(name: str):
    dname = darwin_name(name)
    mem = shared_memory.SharedMemory(dname)

    try:
        b = bytes(mem.buf[:1])
        print(int(b[0]))
    finally:
        mem.close()
        resource_tracker.unregister(mem._name, "shared_memory")

def test_convert():
    bts = "abcdefgh".encode()
    bigi = int.from_bytes(bts, "big", signed=True)
    print(bigi)

def main():
    # test_sm()
    test_read0("RWTest")
    # test_os()
    # test_convert()

if __name__ == "__main__":
    main()


`mac benchmark write`

```
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^BenchmarkDepthWrite$ hf-utils/sharedmemory -count=1

goos: darwin
goarch: amd64
pkg: hf-utils/sharedmemory
cpu: Intel(R) Core(TM) i5-7267U CPU @ 3.10GHz
=== RUN   BenchmarkDepthWrite
BenchmarkDepthWrite
=== RUN   BenchmarkDepthWrite/Chunk=32
BenchmarkDepthWrite/Chunk=32
BenchmarkDepthWrite/Chunk=32-4            692799              2762 ns/op              14 B/op          0 allocs/op
=== RUN   BenchmarkDepthWrite/Chunk=64
BenchmarkDepthWrite/Chunk=64
BenchmarkDepthWrite/Chunk=64-4            742755              1561 ns/op               7 B/op          0 allocs/op
=== RUN   BenchmarkDepthWrite/Chunk=128
BenchmarkDepthWrite/Chunk=128
BenchmarkDepthWrite/Chunk=128-4           884709              1332 ns/op               4 B/op          0 allocs/op
PASS
ok      hf-utils/sharedmemory   4.820s
```

`mac benchmark read`

```
Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^BenchmarkDepthRead$ hf-utils/sharedmemory -count=1

goos: darwin
goarch: amd64
pkg: hf-utils/sharedmemory
cpu: Intel(R) Core(TM) i5-7267U CPU @ 3.10GHz
=== RUN   BenchmarkDepthRead
BenchmarkDepthRead
=== RUN   BenchmarkDepthRead/Chunk=32
BenchmarkDepthRead/Chunk=32
BenchmarkDepthRead/Chunk=32-4             620732              1810 ns/op            3342 B/op          0 allocs/op
=== RUN   BenchmarkDepthRead/Chunk=64
BenchmarkDepthRead/Chunk=64
BenchmarkDepthRead/Chunk=64-4             630672              1704 ns/op            3335 B/op          0 allocs/op
=== RUN   BenchmarkDepthRead/Chunk=128
BenchmarkDepthRead/Chunk=128
BenchmarkDepthRead/Chunk=128-4            625123              1600 ns/op            3267 B/op          0 allocs/op
PASS
ok      hf-utils/sharedmemory   4.692s
```

`172.16.20.91 benchmark write`

```
Running tool: /snap/bin/go test -benchmem -run=^$ -bench ^BenchmarkDepthWrite$ hf-utils/sharedmemory

goos: linux
goarch: amd64
pkg: hf-utils/sharedmemory
cpu: Intel(R) Xeon(R) CPU E5-2683 v3 @ 2.00GHz
=== RUN   BenchmarkDepthWrite
BenchmarkDepthWrite
=== RUN   BenchmarkDepthWrite/Chunk=32
BenchmarkDepthWrite/Chunk=32
BenchmarkDepthWrite/Chunk=32-56                   540668              2049 ns/op              14 B/op          0 allocs/op
=== RUN   BenchmarkDepthWrite/Chunk=64
BenchmarkDepthWrite/Chunk=64
BenchmarkDepthWrite/Chunk=64-56                   640957              1742 ns/op               7 B/op          0 allocs/op
=== RUN   BenchmarkDepthWrite/Chunk=128
BenchmarkDepthWrite/Chunk=128
BenchmarkDepthWrite/Chunk=128-56                  761534              1704 ns/op               4 B/op          0 allocs/op
PASS
ok      hf-utils/sharedmemory   6.129s
```


`172.16.20.91 benchmark read`

```
Running tool: /snap/bin/go test -benchmem -run=^$ -bench ^BenchmarkDepthRead$ hf-utils/sharedmemory

goos: linux
goarch: amd64
pkg: hf-utils/sharedmemory
cpu: Intel(R) Xeon(R) CPU E5-2683 v3 @ 2.00GHz
=== RUN   BenchmarkDepthRead
BenchmarkDepthRead
=== RUN   BenchmarkDepthRead/Chunk=32
BenchmarkDepthRead/Chunk=32
BenchmarkDepthRead/Chunk=32-56            358995              3296 ns/op            3342 B/op          0 allocs/op
=== RUN   BenchmarkDepthRead/Chunk=64
BenchmarkDepthRead/Chunk=64
BenchmarkDepthRead/Chunk=64-56            447122              2641 ns/op            3335 B/op          0 allocs/op
=== RUN   BenchmarkDepthRead/Chunk=128
BenchmarkDepthRead/Chunk=128
BenchmarkDepthRead/Chunk=128-56           445171              2490 ns/op            3267 B/op          0 allocs/op
PASS
ok      hf-utils/sharedmemory   3.585s
```
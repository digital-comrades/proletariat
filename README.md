# proletariat

![Go](https://github.com/digital-comrades/proletariat/workflows/Go/badge.svg)

Simple message passing between systems. This primitive is not fault-tolerant, 
and can be used as a base message passing to build more reliable primitives.
The network connection is done using simple TCP connections based on the `net`
package. The only guarantees are the ones implemented by TCP itself and nothing more.

### Benchmark

A simple benchmark, creates a sever and client and publishes 1024 messages:

```text
goos: linux
goarch: amd64
pkg: github.com/digital-comrades/proletariat/test
cpu: AMD Ryzen 5 3500X 6-Core Processor             
Benchmark_CommunicationMessages             	       1	1003044424 ns/op	  487976 B/op	   13564 allocs/op
Benchmark_CommunicationMessages-2           	       1	1005693788 ns/op	  399464 B/op	   13504 allocs/op
Benchmark_CommunicationMessages-3           	       1	1005936827 ns/op	  403128 B/op	   13498 allocs/op
Benchmark_CommunicationMessages-6           	       1	1006335248 ns/op	  393024 B/op	   13502 allocs/op
Benchmark_CommunicationParallelMessages     	       1	1005231383 ns/op	  400608 B/op	   13521 allocs/op
Benchmark_CommunicationParallelMessages-2   	       1	1005472590 ns/op	  401072 B/op	   13504 allocs/op
Benchmark_CommunicationParallelMessages-3   	       1	1006118222 ns/op	  409912 B/op	   13520 allocs/op
Benchmark_CommunicationParallelMessages-6   	       1	1005206180 ns/op	  404376 B/op	   13529 allocs/op
```

### Comments

This is a simple library to wraps any boilerplate needed when using the `net` package, providing 
the most basic features and nothing fancy. This is packed using the encoder/decoder maintained by
the Hashicorp people, had some trouble when reading directly from the `bufio` so it was easier to
add a decoder to handle this, but this could be a TODO for the future.

Also added a connection pool when dealing with writes, using always a single connection could lead
to writes always failing with `short write`. Using a pool of connections seems to fix this issue, since
different connections can be used to send messages to a destination, but the client should be careful to
not spawn too many connections or a `too many open files` error can happen. Some improvements also could
be made here, handle idle connections, create a pool of workers for handling writes are more high level
and simple to develop, and also try something different like using `io_uring` or something similar at a
low level syscalls.

# proletariat

![Go](https://github.com/digital-comrades/proletariat/workflows/Go/badge.svg)

Simple message passing between systems. This primitive is not fault-tolerant, 
and can be used as a base message passing to build more reliable primitives.
The network connection is done using simple TCP connections based on the `net`
package. The only guarantees are the ones implemented by TCP itself and nothing more.

### Benchmark

A simple benchmark, creates a sever and client for listening messages and publishes 1024 messages:

```text
goos: linux
goarch: amd64
pkg: github.com/digital-comrades/proletariat/test
cpu: AMD Ryzen 5 3500X 6-Core Processor             
Benchmark_CommunicationMessages             	     325	   3797034 ns/op	  371058 B/op	   13248 allocs/op
Benchmark_CommunicationMessages-2           	     746	   2569464 ns/op	  284241 B/op	   10190 allocs/op
Benchmark_CommunicationMessages-3           	    1527	   2282558 ns/op	  353102 B/op	   12626 allocs/op
Benchmark_CommunicationMessages-6           	    1669	   1212951 ns/op	  333823 B/op	   11945 allocs/op
Benchmark_CommunicationParallelMessages     	     403	   3714202 ns/op	  363654 B/op	   12990 allocs/op
Benchmark_CommunicationParallelMessages-2   	     786	   1745899 ns/op	  178969 B/op	    6469 allocs/op
Benchmark_CommunicationParallelMessages-3   	     952	   1864454 ns/op	  193721 B/op	    6991 allocs/op
Benchmark_CommunicationParallelMessages-6   	    1334	   2575388 ns/op	  271032 B/op	    9719 allocs/op
```

### Comments

This is a simple library to wraps any boilerplate needed when using the `net` package, providing 
the most basic features and nothing fancy. This is packed using the encoder/decoder, had some trouble 
when reading directly from the `bufio` so it was easier to add a decoder to handle this, but this could 
be a TODO for the future.

Also added a connection pool when dealing with writes, using always a single connection could lead
to writes always failing with `short write`. Using a pool of connections seems to fix this issue, since
different connections can be used to send messages to a destination, but the client should be careful to
not spawn too many connections or a `too many open files` error can happen. Some improvements also could
be made here, handle idle connections, create a pool of workers for handling writes are more high level
and simple to develop, and also try something different like using `io_uring` or something similar at a
low level syscalls.

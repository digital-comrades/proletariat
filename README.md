# proletariat

Simple message passing between systems. This primitive is not fault-tolerant, 
this can be used as a base message passing to build more reliable primitives.

## Available

Will be available for the client the possibility to choose some basic parameters,
like the bind address, the timeouts, TCP/UDP connections. Every call will be made over
the TCP/IP stack, even on a local network, and the primitives are **not** reliable, will
work only with the best effort.

When the system start, the `Communication` will bind to the configured address and start
polling for data. Every call made to send messages will be non-blocking, and a response is not
required to return. Which means this is not a request-response model, if the system `A` send
a message to system `B` and a response needs to be sent back, the system `B` must have a side effect
to call the system `A`. 

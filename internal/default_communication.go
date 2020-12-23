package internal

import (
	"context"
	"github.com/jabolina/proletariat/pkg/proletariat"
	"net"
	"sync"
	"time"
)

const (
	minPollDelay = 5 * time.Millisecond
	maxPollDelay = 500 * time.Millisecond
)

// Default struct that implements the Communication interface.
// Using this implementation is possible to send and receive messages.
type DefaultCommunication struct {
	// Synchronize operations on available connections.
	mutex *sync.Mutex

	// Handler to carefully invoke new goroutines.
	handler *GoRoutineHandler

	// Configuration for the communication primitive.
	configuration proletariat.CommunicationConfiguration

	// Transport used to send and receive messages.
	transport proletariat.Transport

	// Channel that will receive data from another connections.
	listener chan []byte

	// All established connections.
	connections map[proletariat.Address]proletariat.Connection

	localCtx context.Context

	cancelCtx context.CancelFunc
}

func NewCommunication(configuration proletariat.CommunicationConfiguration) proletariat.Communication {
	ctx, cancel := context.WithCancel(context.TODO())
	return &DefaultCommunication{
		mutex:         &sync.Mutex{},
		handler:       NewRoutineHandler(),
		configuration: configuration,
		transport:     configuration.Transport,
		listener:      make(chan []byte),
		connections:   make(map[proletariat.Address]proletariat.Connection),
		localCtx:      ctx,
		cancelCtx:     cancel,
	}
}

// When a new connection request is received by the server this method is
// initiated. Using the given net connection a wrapper is created for this
// incoming request.
// This incoming connection request, will remain open until the peer closes,
// polling and for every received data will publish to the listener channel.
// After created, since this connection is created only to receive messages
// it will not be stored for the in-memory connections.
func (d *DefaultCommunication) handleIncomingConnection(conn net.Conn) {
	select {
	case <-d.localCtx.Done():
		return
	case <-d.configuration.Ctx.Done():
		return
	default:
		ctx, cancel := context.WithCancel(d.configuration.Ctx)
		incoming := ConnectionConfiguration{
			Timeout:    d.configuration.Timeout,
			Read:       d.listener,
			Ctx:        ctx,
			Connection: conn,
			Target:     proletariat.Address(conn.RemoteAddr().String()),
		}
		connection := NewNetworkConnection(incoming)
		d.handler.Spawn(connection.Listen)

		<-d.configuration.Ctx.Done()
		cancel()
	}
}

// For a given address, create a new connection instance if possible.
func (d *DefaultCommunication) resolveConnection(address proletariat.Address) (proletariat.Connection, error) {
	if connection := d.getActiveConnection(address); connection != nil {
		return connection, nil
	}
	return d.establishNewConnection(address)
}

// Retrieve a connection for the in-memory available connections.
func (d *DefaultCommunication) getActiveConnection(address proletariat.Address) proletariat.Connection {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.connections[address]
}

// Establish a connection with another peer using the available transport if possible.
func (d *DefaultCommunication) establishNewConnection(address proletariat.Address) (proletariat.Connection, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	conn, err := d.transport.Bind(address, d.configuration.Timeout)
	if err != nil {
		return nil, err
	}
	config := ConnectionConfiguration{
		Timeout:    d.configuration.Timeout,
		Read:       d.listener,
		Connection: conn,
		Target:     address,
	}
	connection := NewNetworkConnection(config)
	d.connections[address] = connection
	return connection, nil
}

func (d *DefaultCommunication) acceptConnections() error {
	conn, err := d.transport.Accept()
	if err != nil {
		return err
	}
	d.handler.Spawn(func() {
		d.handleIncomingConnection(conn)
	})
	return nil
}

func (d *DefaultCommunication) poll() {
	var pollDelay = minPollDelay
	for {
		pollDelay = min(pollDelay*2, maxPollDelay)
		err := d.acceptConnections()
		if err == nil {
			pollDelay = minPollDelay
		}

		select {
		case <-d.localCtx.Done():
			return
		case <-d.configuration.Ctx.Done():
			return
		case <-time.After(pollDelay):
			continue
		}
	}
}

// Implements the Communication interface.
func (d *DefaultCommunication) Close() error {
	d.cancelCtx()
	d.mutex.Lock()
	defer d.mutex.Unlock()
	for _, connection := range d.connections {
		if err := connection.Close(); err != nil {
			return err
		}
	}
	d.handler.Close()
	return d.transport.Close()
}

// Implements the Communication interface.
func (d *DefaultCommunication) Start() {
	go d.poll()
}

// Implements the Communication interface.
func (d *DefaultCommunication) Send(address proletariat.Address, data []byte) error {
	connection, err := d.resolveConnection(address)
	if err != nil {
		return err
	}
	return connection.Write(data)
}

// Implements the Communication interface.
func (d *DefaultCommunication) Receive() <-chan []byte {
	return d.listener
}

func min(a time.Duration, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

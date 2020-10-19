package grpc_connection_pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// ErrExceeded is the error when maximum number of connections exceeded
var ErrExceeded = errors.New("Maximum number of connections exceeded")
var ErrUnavailable = errors.New("No available connection")

type GRPCPool struct {
	host               string
	options            *Options
	dialOptions        []grpc.DialOption
	connections        chan *Connection
	connCount          uint32
	streamInitializers sync.Map
}

// NewGRPCPool creates a new connection pool.
func NewGRPCPool(host string, options *Options, dialOptions ...grpc.DialOption) (*GRPCPool, error) {

	pool := &GRPCPool{
		host:        host,
		options:     options,
		dialOptions: dialOptions,
		connections: make(chan *Connection, options.MaxCap),
		connCount:   0,
	}

	err := pool.init()
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (pool *GRPCPool) init() error {

	// Initializing connections
	for i := 0; i < pool.options.InitCap; i++ {

		// Create connection
		connection, err := pool.factory()
		if err != nil {
			return err
		}

		pool.connections <- NewConnection(pool, connection)
	}

	return nil
}

func (pool *GRPCPool) getConnectionCount() uint32 {
	return atomic.LoadUint32((*uint32)(&pool.connCount))
}

func (pool *GRPCPool) ref() (uint32, error) {

	count := pool.getConnectionCount()

	// Check pool size
	if count >= uint32(pool.options.MaxCap) {
		return count, ErrExceeded
	}

	// Update counter
	count = atomic.AddUint32((*uint32)(&pool.connCount), 1)

	return count, nil
}

func (pool *GRPCPool) unref() (uint32, error) {
	return atomic.AddUint32((*uint32)(&pool.connCount), ^uint32(0)), nil
}

func (pool *GRPCPool) factory() (*grpc.ClientConn, error) {

	_, err := pool.ref()
	if err != nil {
		return nil, ErrExceeded
	}

	// Preparing context with timeout options
	ctx, cancel := context.WithTimeout(context.Background(), pool.options.DialTimeout)
	defer cancel()

	connection, err := grpc.DialContext(ctx, pool.host, pool.dialOptions...)
	if err != nil {
		pool.unref()
		return nil, err
	}

	return connection, nil
}

func (pool *GRPCPool) checkConnectionState(connection *grpc.ClientConn) bool {

	state := connection.GetState()

	if state == connectivity.Shutdown || state == connectivity.TransientFailure {

		// this connection doesn't work
		connection.Close()
		pool.unref()

		return false
	}

	return true
}

func (pool *GRPCPool) get() (*Connection, error) {

	for {

		select {
		case c := <-pool.connections:

			// Getting connection from buffered channel
			if !pool.checkConnectionState(c.connection) {
				continue
			}

			// Put connection back to pool immediately
			pool.connections <- c

			return c, nil
		default:

			// No available connection, so creating a new connection
			c, err := pool.factory()
			if err != nil {

				// Cannnot establish more connection
				if err == ErrExceeded {
					continue
				}

				if pool.getConnectionCount() == uint32(0) {
					// No available connection
					return nil, ErrUnavailable
				}

				continue
			}

			pool.Push(c)
		}
	}
}

func (pool *GRPCPool) pop() (*Connection, error) {

	for {

		select {
		case c := <-pool.connections:

			// Getting connection from buffered channel
			if !pool.checkConnectionState(c.connection) {
				continue
			}

			return c, nil
		default:

			// No available connection, so creating a new connection
			c, err := pool.factory()
			if err != nil {

				// Cannnot establish more connection
				if err == ErrExceeded {
					continue
				}

				if pool.getConnectionCount() == uint32(0) {
					// No available connection
					return nil, ErrUnavailable
				}
			}

			pool.Push(c)
		}
	}
}

func (pool *GRPCPool) push(connection *Connection) error {

	if !pool.checkConnectionState(connection.connection) {
		return nil
	}

	pool.connections <- connection

	return nil
}

// Get will returns a available gRPC client.
func (pool *GRPCPool) Get() (*grpc.ClientConn, error) {

	conn, err := pool.get()
	if err != nil {
		return nil, err
	}

	return conn.connection, nil
}

// Pop will return a availabe gRPC client and the gRPC client will not be reused before return client to the pool.
func (pool *GRPCPool) Pop() (*grpc.ClientConn, error) {

	conn, err := pool.pop()
	if err != nil {
		return nil, err
	}

	return conn.connection, nil
}

// Push will put gRPC client to the pool.
func (pool *GRPCPool) Push(connection *grpc.ClientConn) error {
	return pool.push(NewConnection(pool, connection))
}

package grpc_connection_pool

import (
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type StreamInitializer func(*grpc.ClientConn) (interface{}, error)

type Connection struct {
	pool        *GRPCPool
	connection  *grpc.ClientConn
	updatedTime time.Time
	streams     *sync.Map
}

func NewConnection(pool *GRPCPool, c *grpc.ClientConn) *Connection {
	return &Connection{
		pool:        pool,
		connection:  c,
		updatedTime: time.Now(),
		streams:     &sync.Map{},
	}
}

func (connection *Connection) GetStream(name string) (interface{}, error) {

	// Getting stream by connection
	val, ok := connection.streams.Load(name)
	if ok {
		return val, nil
	}

	// Initialize stream for connection
	initializer := connection.pool.GetStreamInitializer(name)
	if initializer == nil {
		return nil, errors.New("No such stream initializer")
	}

	stream, err := initializer(connection.connection)
	if err != nil {
		return nil, err
	}

	connection.streams.Store(name, stream)

	return stream, nil
}

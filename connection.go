package grpc_connection_pool

import (
	"errors"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type StreamInitializer func(*grpc.ClientConn) (interface{}, error)
type StreamHandler func(interface{}) error

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

func (connection *Connection) GetStream(name string, fn StreamHandler) error {

	// Getting stream by connection
	val, ok := connection.streams.Load(name)
	if ok {
		err := fn(val)
		if err == io.EOF {
			connection.streams.Delete(name)
			return connection.GetStream(name, fn)
		}

		if err != nil {
			return err
		}
	}

	// Initialize stream for connection
	initializer := connection.pool.GetStreamInitializer(name)
	if initializer == nil {
		return errors.New("No such stream initializer")
	}

	stream, err := initializer(connection.connection)
	if err != nil {
		return err
	}

	if stream == nil {
		return errors.New("Invalid stream")
	}

	err = fn(stream)
	if err != io.EOF {
		connection.streams.Store(name, stream)
	}

	return err
}

package grpc_connection_pool

import (
	"time"

	"google.golang.org/grpc"
)

type Connection struct {
	connection  *grpc.ClientConn
	updatedTime time.Time
}

func NewConnection(c *grpc.ClientConn) *Connection {
	return &Connection{
		connection:  c,
		updatedTime: time.Now(),
	}
}

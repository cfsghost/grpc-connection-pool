package grpc_connection_pool

import "time"

// Options represent all of the available options when creating a connection pool with NewGRPCPool.
type Options struct {
	InitCap     int
	MaxCap      int
	DialTimeout time.Duration
}

// NewOptions creates a Options object.
func NewOptions() *Options {
	return &Options{
		InitCap:     8,
		MaxCap:      128,
		DialTimeout: 10 * time.Second,
	}
}

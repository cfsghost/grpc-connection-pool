/*

Package grpc_connection_pool provides the ability to create connection pool for gRPC.

Here is example to create connection pool:

	// Create Options object
	options := &pool.Options{
		InitCap: 8,
		MaxCap: 16,
		DialTimeout: time.Second * 20,
	}

	// Create connection pool with options
	pool, err := grpc_connection_pool.NewGRPCPool("localhost:8888", options, grpc.WithInsecure())
	if err != nil {
		log.Println(err)
		return
	}

	// Get a available gRPC client
	client := pool.Get()

*/
package grpc_connection_pool

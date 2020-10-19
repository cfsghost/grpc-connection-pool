package grpc_connection_pool

func (pool *GRPCPool) SetStreamInitializer(name string, initializer StreamInitializer) {
	pool.streamInitializers.Store(name, initializer)
}

func (pool *GRPCPool) GetStreamInitializer(name string) StreamInitializer {
	initializer, ok := pool.streamInitializers.Load(name)
	if !ok {
		return nil
	}

	return initializer.(StreamInitializer)
}

func (pool *GRPCPool) GetStream(name string, fn StreamHandler) error {

	conn, err := pool.get()
	if err != nil {
		return err
	}

	return conn.GetStream(name, fn)
}

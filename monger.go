package monger

func newConfigs(opts ...ConfigOption) *Config {
	config := Config{
		Hosts: []string{"localhost"},
	}

	for _, o := range opts {
		o(&config)
	}

	return &config
}

/*
Connect is the function to connect MongoDB with mgo.v2 deiver
*/
func Connect(opts ...ConfigOption) (Connection, error) {
	config := newConfigs(opts...)
	// panic("The function has not Impl")

	conn := newConnection(config, nil)

	if err := conn.Open(); err != nil {
		return nil, err
	}

	return conn, nil
}

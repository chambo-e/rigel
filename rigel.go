package rigel

import redis "gopkg.in/redis.v5"

// Rigel is client struct
type Rigel struct {
	redis     *redis.Client
	namespace string
}

// Config represents Rigel configuration
type Config struct {
	// Redis is the remote redis configuration https://godoc.org/gopkg.in/redis.v5#Options
	// If it's empty it will connect to redis://localhost:6379 by default
	Redis redis.Options
	// Namespace is the namespace that needs to be used to isolates keys and publish messages on redis.
	// Default is "rigel"
	Namespace string
}

// New instanciate a new Rigel client that can be used to process or enqueue jobs
func New(config Config) (*Rigel, error) {
	if config.Redis.Addr == "" {
		config.Redis.Addr = "localhost:6379"
	}

	if config.Namespace == "" {
		config.Namespace = "rigel"
	}

	client := redis.NewClient(&config.Redis)
	if err := client.Ping().Err(); err != nil {
		return nil, err
	}

	return &Rigel{
		redis:     client,
		namespace: config.Namespace,
	}, nil
}

package rigel

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	configs := []struct {
		config     Config
		shouldFail bool
	}{
		{Config{}, false},
		{Config{Namespace: "test"}, false},
		{Config{Namespace: "test", Concurrency: 10}, false},
		{Config{Namespace: "test", Concurrency: 10, Hostname: "test"}, false},
		{Config{Namespace: "test", Concurrency: 10, Hostname: "test", Redis: RedisOptions{}}, false},
		{Config{Namespace: "test", Concurrency: 10, Hostname: "test", Redis: RedisOptions{Addr: "hello:6666", DialTimeout: time.Second}}, true},
	}

	for _, x := range configs {
		r, err := New(x.config)

		if x.shouldFail {
			assert.NotNil(t, err, "err should not be nil")
			return
		}

		assert.Nil(t, err, "err should be nil")

		if x.config.Concurrency == 0 {
			x.config.Concurrency = 1
		}
		if x.config.Namespace == "" {
			x.config.Namespace = "rigel"
		}
		if x.config.Hostname == "" {
			h, _ := os.Hostname()
			x.config.Hostname = h
		}
		if x.config.Redis.Addr == "" {
			x.config.Redis.Addr = "localhost:6379"
		}

		assert.Equal(t, x.config.Concurrency, r.concurrency, "Concurrency setting is not correctly setted")
		assert.Equal(t, x.config.Namespace, r.namespace, "Namespace setting is not correctly setted")
		assert.Equal(t, x.config.Hostname, r.hostname, "Hostname setting is not correctly setted")
	}
}

func TestRigel_Handle(t *testing.T) {
	r, err := New(Config{})
	assert.Nil(t, err, "err should be nil")

	funcs := []struct {
		queue string
		count int
		fn    Handler
	}{
		{"test", 1, func([]byte) error {
			return nil
		}},
		{"test", 1, func([]byte) error {
			return nil
		}},
		{"titi", 2, func([]byte) error {
			return nil
		}},
		{"toto", 3, func([]byte) error {
			return nil
		}},
		{"toto", 3, func([]byte) error {
			return nil
		}},
	}

	for _, x := range funcs {
		r.Handle(x.queue, x.fn)

		assert.Equal(t, len(r.handlers), x.count, "Amount of handlers and count should be equal")
		assert.Contains(t, r.handlers, x.queue, "handlers should contains '"+x.queue+"'handler")
	}
}

func TestRigel_Start(t *testing.T) {
}

func TestRigel_Stop(t *testing.T) {
}

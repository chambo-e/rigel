package rigel

import (
	"fmt"

	redis "gopkg.in/redis.v5"
)

// Enqueue enqueue job into queue q
func (r *Rigel) Enqueue(q string, job []byte) error {
	if _, err := r.redis.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.SAdd(fmt.Sprintf("%s:queues", r.namespace), q)
		pipe.RPush(fmt.Sprintf("%s:queue:%s", r.namespace, q), job)
		pipe.Publish(fmt.Sprintf("%s:queue:%s", r.namespace, q), string(job))
		return nil
	}); err != nil {
		return err
	}
	return nil
}

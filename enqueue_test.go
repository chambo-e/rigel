package rigel

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRigel_Enqueue(t *testing.T) {
	r, err := New(Config{})
	assert.Nil(t, err, "err should be nil")
	defer r.redis.FlushAll()

	jobs := []struct {
		queue string
		job   []byte
		count int64
	}{
		{"test", []byte("hello"), 1},
		{"test", []byte("hello"), 2},
		{"test_1", []byte("hello"), 1},
		{"test_2", []byte("hello"), 1},
		{"test_3", []byte("hello"), 1},
		{"test_4", []byte("hello"), 1},
	}

	for x := 0; x < 100; x++ {
		jobs = append(jobs, struct {
			queue string
			job   []byte
			count int64
		}{"test_load", []byte("hello"), int64(x + 1)})
	}

	for _, x := range jobs {
		ch := make(chan bool)

		go func() {
			sub, err := r.redis.Subscribe(fmt.Sprintf("%s:queue:%s", r.namespace, x.queue))
			assert.Nil(t, err, "err should be nil")
			defer sub.Close()

			// Default timeout is 5s
			if _, err := sub.ReceiveMessage(); err != nil {
				ch <- false
				return
			}
			ch <- true
		}()

		<-time.After(150 * time.Millisecond)
		r.Enqueue(x.queue, x.job)

		assert.Equal(t, true, <-ch, "should have received a msg from redis subscription")
		assert.Equal(t, x.count, r.redis.LLen(fmt.Sprintf("%s:queue:%s", r.namespace, x.queue)).Val(), "mismatch in amount of job stored in queue")
		assert.Contains(t, r.redis.SMembers(fmt.Sprintf("%s:queues", r.namespace)).Val(), x.queue, "queues set should contains queue name")
		close(ch)
	}
}

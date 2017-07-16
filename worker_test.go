package rigel

import (
	"fmt"
	"testing"
	"time"

	redis "gopkg.in/redis.v5"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestWorker_Start(t *testing.T) {
	r, err := New(Config{})
	assert.Nil(t, err, "err should be nil")
	defer r.redis.FlushAll()

	pipe := make(chan _Job)
	defer close(pipe)
	wrkr := r.newWorker()

	ch := make(chan *redis.Message)
	defer close(ch)

	go func() {
		sub, er := r.redis.PSubscribe(fmt.Sprintf("%s:worker:*", wrkr.namespace))
		assert.Nil(t, er, "err should be nil")
		defer sub.Close()

		for x := 0; x < 2; x++ {
			// Default timeout is 5s
			msg, er := sub.ReceiveMessage()
			if er != nil {
				ch <- nil
				return
			}
			ch <- msg
		}
	}()

	<-time.After(150 * time.Millisecond)
	go wrkr.Start(pipe)

	msg := <-ch
	tStart, err := time.Parse(time.RFC3339, msg.Payload)
	assert.Nil(t, err, "err should be nil")
	spew.Dump(msg.Payload)
	spew.Dump(tStart)
	assert.Equal(t, fmt.Sprintf("%s:worker:%s:%s:started", wrkr.namespace, wrkr.hostname, wrkr.id), msg.Channel, "")
	assert.WithinDuration(t, time.Now(), tStart, time.Second, "")

	msg = <-ch
	assert.Equal(t, fmt.Sprintf("%s:worker:%s:%s", wrkr.namespace, wrkr.hostname, wrkr.id), msg.Channel, "")
	assert.Equal(t, _STARTING, msg.Payload, "")

	// assert.Equal(t, , <-ch, "should have received a msg from redis subscription")
	// assert.Equal(t, x.count, r.redis.LLen(fmt.Sprintf("%s:queue:%s", r.namespace, x.queue)).Val(), "mismatch in amount of job stored in queue")
	// assert.Contains(t, r.redis.SMembers(fmt.Sprintf("%s:queues", r.namespace)).Val(), x.queue, "queues set should contains queue name")

}

func TestWorker_Stop(t *testing.T) {
}

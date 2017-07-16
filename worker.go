package rigel

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/oklog/ulid"

	redis "gopkg.in/redis.v5"
)

const (
	_STARTING = "STARTING"
	_POLLING  = "POLLING"
	_WORKING  = "WORKING"
	_EXITING  = "EXITING"
	_EXITED   = "EXITED"
)

type worker struct {
	id string

	namespace string
	hostname  string
	redis     *redis.Client

	handlers map[string]Handler

	wg   sync.WaitGroup
	quit chan struct{}
}

func (r *Rigel) newWorker() *worker {
	t := time.Now()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	idU, err := ulid.New(ulid.Timestamp(t), entropy)
	var id string
	if err != nil {
		id = strconv.FormatInt(rand.Int63n(math.MaxUint64%2)+time.Now().UnixNano(), 10)
	} else {
		id = idU.String()
	}

	return &worker{
		id:        id,
		namespace: r.namespace,
		hostname:  r.hostname,
		redis:     r.redis,
		handlers:  r.handlers,
		quit:      make(chan struct{}),
	}
}

// Start starts the worker
// It will run until Stop() is called
func (w *worker) Start(pipe <-chan _Job) {
	w.wg.Add(1)
	w.broadcastStart()

L:
	for {
		w.setState(_POLLING)
		select {
		case <-w.quit:
			w.setState(_EXITING)
			break L
		case job := <-pipe:
			w.broadcastWorking(job)
			w.handle(job)
			w.clearJob()
		}
	}

	w.setState(_EXITED)
	w.cleanup()
	w.wg.Done()
}

func (w *worker) broadcastStart() {
	if _, err := w.redis.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.SAdd(fmt.Sprintf("%s:workers", w.namespace), fmt.Sprintf("%s:%s", w.hostname, w.id))
		pipe.HMSet(fmt.Sprintf("%s:worker:%s:%s", w.namespace, w.hostname, w.id), map[string]string{"started": time.Now().Format(time.RFC3339)})
		pipe.Publish(fmt.Sprintf("%s:worker:%s:%s:started", w.namespace, w.hostname, w.id), time.Now().Format(time.RFC3339))

		// Set state to STARTED
		// Done here instead of setState to take advantage of the current MULTI EXEC
		pipe.HMSet(fmt.Sprintf("%s:worker:%s:%s", w.namespace, w.hostname, w.id), map[string]string{"state": _STARTING})
		pipe.Publish(fmt.Sprintf("%s:worker:%s:%s:state", w.namespace, w.hostname, w.id), _STARTING)
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (w *worker) setState(state string) {
	if _, err := w.redis.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.HMSet(fmt.Sprintf("%s:worker:%s:%s", w.namespace, w.hostname, w.id), map[string]string{"state": state})
		pipe.Publish(fmt.Sprintf("%s:worker:%s:%s:state", w.namespace, w.hostname, w.id), state)
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (w *worker) broadcastWorking(job _Job) {
	if _, err := w.redis.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.HMSet(fmt.Sprintf("%s:worker:%s:%s", w.namespace, w.hostname, w.id), map[string]string{"job": string(job.Job)})
		pipe.Publish(fmt.Sprintf("%s:worker:%s:%s:working", w.namespace, w.hostname, w.id), string(job.Job))

		// Set state to WORKING
		// Done here instead of setState to take advantage of the current MULTI EXEC
		pipe.HMSet(fmt.Sprintf("%s:worker:%s:%s", w.namespace, w.hostname, w.id), map[string]string{"state": _WORKING})
		pipe.Publish(fmt.Sprintf("%s:worker:%s:%s:state", w.namespace, w.hostname, w.id), _WORKING)
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (w *worker) handle(job _Job) {
	w.incrStat("processed", job.Queue)
	if err := w.handlers[job.Queue](job.Job); err != nil {
		w.enqueueFailed(job, err)
		w.incrStat("failed", job.Queue)
		return
	}
	w.incrStat("succeeded", job.Queue)
}

func (w *worker) incrStat(stat, q string) {
	if _, err := w.redis.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.HIncrBy(fmt.Sprintf("%s:worker:%s:%s", w.namespace, w.hostname, w.id), stat, 1)
		pipe.Publish(fmt.Sprintf("%s:worker:%s:%s:%s", w.namespace, w.hostname, w.id, stat), "1")
		pipe.HIncrBy(fmt.Sprintf("%s:queues:%s:stats", w.namespace, q), stat, 1)
		pipe.Publish(fmt.Sprintf("%s:queues:%s:stats:%s", w.namespace, q, stat), "1")
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (w *worker) enqueueFailed(job _Job, err error) {
	failed, err := json.Marshal(_FailedJob{
		Job:      string(job.Job),
		FailedAt: time.Now(),
		Err:      err.Error(),
	})
	if err != nil {
		log.Println(err)
		return
	}

	if _, err := w.redis.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.RPush(fmt.Sprintf("%s:queue:%s:failed", w.namespace, job.Queue), failed)
		pipe.Publish(fmt.Sprintf("%s:queue:%s:failed", w.namespace, job.Queue), string(failed))
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (w *worker) clearJob() {
	if _, err := w.redis.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.HDel(fmt.Sprintf("%s:worker:%s:%s", w.namespace, w.hostname, w.id), "job")
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (w *worker) cleanup() {
	if _, err := w.redis.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.Del(fmt.Sprintf("%s:worker:%s:%s", w.namespace, w.hostname, w.id))
		pipe.SRem(fmt.Sprintf("%s:workers", w.namespace), fmt.Sprintf("%s:%s", w.hostname, w.id))
		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (w *worker) Stop() {
	close(w.quit)
	w.wg.Wait()
}

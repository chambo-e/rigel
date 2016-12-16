package rigel

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	redis "gopkg.in/redis.v5"
)

// Rigel is client struct
type Rigel struct {
	redis       *redis.Client
	namespace   string
	concurrency int
	hostname    string

	handlers map[string]Handler
	queues   map[string]bool
	workers  []*worker

	quit chan struct{}
	wg   sync.WaitGroup
	mtx  sync.RWMutex
}

// Config represents Rigel configuration
type Config struct {
	// Redis is the remote redis configuration https://godoc.org/gopkg.in/redis.v5#Options
	// If it's empty it will connect to redis://localhost:6379 by default
	Redis redis.Options
	// Namespace is the namespace that needs to be used to isolates keys and publish messages on redis.
	// Default is "rigel"
	Namespace string
	// Concurrency is the amount of workers to start (default: 1)
	Concurrency int
	// Hostname of the host (default: os.Hostname() or "")
	Hostname string
}

// Handler define the function interface used as job handler
// Handler must return an error or nil if processing went without troubles
type Handler func([]byte) error

// New instanciate a new Rigel client that can be used to process or enqueue jobs
func New(config Config) (*Rigel, error) {
	if config.Redis.Addr == "" {
		config.Redis.Addr = "localhost:6379"
	}

	if config.Namespace == "" {
		config.Namespace = "rigel"
	}

	if config.Concurrency == 0 {
		config.Concurrency = 1
	}

	if config.Hostname == "" {
		config.Hostname, _ = os.Hostname()
	}

	client := redis.NewClient(&config.Redis)
	if err := client.Ping().Err(); err != nil {
		return nil, err
	}

	return &Rigel{
		redis:       client,
		namespace:   config.Namespace,
		hostname:    config.Hostname,
		handlers:    make(map[string]Handler),
		queues:      make(map[string]bool),
		concurrency: config.Concurrency,
		quit:        make(chan struct{}),
	}, nil
}

// Handle register a Handler processing jobs from q
// It must be called before Start()
func (r *Rigel) Handle(q string, handler Handler) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.handlers[q] = handler
	r.queues[q] = true
}

// Start run the poller and start the workers.
// It's blocking until Stop() is called
func (r *Rigel) Start() error {
	pipe, err := r.poll()
	if err != nil {
		return err
	}
	defer close(pipe)

	for x := 0; x < r.concurrency; x++ {
		r.wg.Add(1)
		wrkr := r.newWorker()
		go wrkr.Start(pipe)
		r.mtx.Lock()
		r.workers = append(r.workers, wrkr)
		r.mtx.Unlock()
	}

	r.wg.Wait()
	if len(pipe) > 0 {
		r.requeue(pipe)
	}
	return nil
}

// Stop shut down polling and workers.
// There might be a small delay waiting for workers to complete their current job processing
func (r *Rigel) Stop() {
	close(r.quit)

	r.mtx.Lock()
	defer r.mtx.Unlock()
	for _, w := range r.workers {
		w.Stop()
		r.wg.Done()
	}
}

func (r *Rigel) poll() (chan _Job, error) {
	var queues []string
	for q := range r.queues {
		queues = append(queues, q)
	}

	pipe := make(chan _Job, r.concurrency)

	for _, q := range queues {
		r.wg.Add(1)

		go func(q string) {
			defer r.wg.Done()

			for {
				select {
				case <-r.quit:
					return
				default:
				}

				job, err := r.redis.BLPop(time.Second, fmt.Sprintf("%s:queue:%s", r.namespace, q)).Result()
				if err != nil {
					if err != redis.Nil {
						log.Println(err)
					}
					continue
				}

				jb := _Job{
					Queue: q,
					Job:   []byte(job[1]),
				}

				pipe <- jb
			}
		}(q)
	}

	return pipe, nil
}

func (r *Rigel) requeue(p <-chan _Job) {
	left := []_Job{}
	for x := range p {
		left = append(left, x)
	}

	if _, err := r.redis.Pipelined(func(pipe *redis.Pipeline) error {
		for x := len(left); x > 0; x-- {
			pipe.SAdd(fmt.Sprintf("%s:queues", r.namespace), left[x].Queue)
			pipe.LPush(fmt.Sprintf("%s:queue:%s", r.namespace, left[x].Queue), left[x].Job)
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
}

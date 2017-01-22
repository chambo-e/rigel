# Rigel

Rigel aims to be a fast and efficient task (or job) queue alternative to Celery, Resque, Sidekiq and so much else.

- [Why ?](#why-)
- [Features](features)
- [Example](#example)
- [Contributing](#contributing)
  - [Why Redis ?](#why-redis-)
  - [Why not `insert message_queue_system` ?](#user-content-why-not-insert-message_queue_system-)

## Why ?

Because there is no inexpensive, fast and easily monitorable system available yet.

## Features

As of now Rigel is still a quite basic task queue. It does enqueue and process jobs (obvious right ?).

Feature              | Wanted | Developing | Integrated | Tested | Version
:------------------- | :----: | :--------: | :--------: | :----- | :------
Enqueue              |        |            |     X      |        |
Processing           |        |            |     X      |        |
Crash safe           |   X    |            |            |        |
Network failure safe |   X    |            |            |        |
Scheduled jobs       |   X    |            |            |        |
Batch jobs           |   X    |            |            |        |
WebUI                |        |     X      |            |        |
Realtime Stats       |   X    |            |            |        |

## Example

`consumer.go`

```go
package main

import (
    "github.com/chambo-e/rigel"
    "log"
)

func main() {
    r, _ := rigel.New(rigel.Config{Concurrency: 25})

    r.Handle("failure", func(job []byte) error {
        return errors.New("broken")
    })
    r.Handle("success", func(job []byte) error {
        log.Printf("Received job from 'success' queue: %s", string(job))
        return nil
    })

    r.Start()
}
```

`producer.go`

```go
package main

import (
    "github.com/chambo-e/rigel"
)

func main() {
    r, _ := rigel.New(rigel.Config{Concurrency: 25})

    r.Enqueue("failure", []byte(""))
    r.Enqueue("success", []byte("hello world"))
}
```

## Contributing

Well Rigel is a therapeutic project, originally started to cure my _not enough perfect_ programmer disease but I would be more than thrilled to discuss a feature or merge a PR. Go ahead, tweak it, break it, make it faster :)

### Why Redis ?

> Because it's fast, has barely no memory overhead, allow non destructive inspection, easy to deploy, already a part of a lot of infrastructures, persistent and has perfectly suited data types for task queues.

#### Why not `insert message_queue_system` ?

- **[RabbitMQ](https://www.rabbitmq.com/)**:

> Strong and robust competitor, Celery is built upon it as many others. Written in Erlang it's also very reliable but unfortunately inspection is destructive and as I want Rigel to be heavily monitorable this would have be a major drawback. It also have more memory overhead than Redis.

- **[NATS Streaming](https://github.com/nats-io/nats-streaming-server)**

> NATS Streaming is a recent platform built upon NATS. While very interesting and efficient it's still very young and miss a few features to be suited for task queues.

- **[Disque](https://github.com/antirez/disque)**

> Disque is @antirez response for people using Redis as task queues brokers. Based on a Redis skeleton it exposes a different API dedicated to handle jobs while offering every features that made Redis perfect for this kind of system. But unfortunately it's not yet a "stable" software. But it's not an [abandoned project](https://github.com/antirez/disque/issues/190), I might reconsider it when it's stable.

- **Custom made**

> I also thought and started developing Rigel with a custom made server. Dedicated to handle jobs and exposing an API like Disque but I would never have been able to match my performance and safety requirements so I gave up and chose Redis. Also not closed to reconsider but that would require a lot of investment and time I don't currently have.

- **[Iron](http://iron.io/)/[AWS Lambda](https://aws.amazon.com/fr/lambda/)/...**

> Serverless systems are based on message queues and are a really nice improvement over traditional task queues. They're auto scaling, costless and painless to manage. But they still come expensive compared to on-premise task queues.

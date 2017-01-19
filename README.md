# Rigel

Rigel aims to be a fast and efficient task queue alternative to every frameworks like Celery, Resque, Sidekiq and so much else.

### Why ?

Because there is no inexpensive, fast and easily monitorable system available yet.

### Features

As of now Rigel is still a quite basic task queue. It does enqueue and process jobs (obvious right ?).

### Contributing

Well Rigel is a therapeutic project, originally started to cure my _not enough perfect_ programmer disease but I would be more than thrilled to discuss a feature or merge a PR. Go ahead, tweak it, break it, make it faster :)

#### Why Redis ?

>Because it's fast, has barely no memory overhead, allow non destructive inspection, easy to deploy, already a part of a lot of infrastructures, persistent and has perfectly suited data types for task queues.

##### Why not `insert message_queue_system` ?

 - **[RabbitMQ](https://www.rabbitmq.com/)**:

>Strong and robust competitor, Celery is built upon it as many others. Written in Erlang it's also very reliable but unfortunately inspection is destructive and as I want Rigel to be heavily monitorable this would have be a major drawback. It also have more memory overhead than Redis.

- **[NATS Streaming](https://github.com/nats-io/nats-streaming-server)**

>NATS Streaming is a recent platform built upon NATS. While very interesting and efficient it's still very young and miss a few features to be suited for task queues.

- **[Disque](https://github.com/antirez/disque)**

>Disque is @antirez response for people using Redis as task queues brokers. Based on a Redis skeleton it exposes a different API dedicated to handle jobs while offering every features that made Redis perfect for this kind of system. But unfortunately it's not yet a "stable" software. But it's not an [abandoned project](https://github.com/antirez/disque/issues/190), I might reconsider it when it's stable.

- **Custom made**

>I also thought and started developing Rigel with a custom made server. Dedicated to handle jobs and exposing an API like Disque but I would never have been able to match my performance and safety requirements so I gave up and chose Redis. Also not closed to reconsider but that would require a lot of investment and time I don't currently have.

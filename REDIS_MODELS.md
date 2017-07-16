# Models

`{NS}` stands for namespace

Name                        | Type | Desc
:-------------------------- | :--- | -------------------------------------------------------------
`{NS}:queues`                 | Set  | Contains names of active queues
`{NS}:queue:{QUEUE}`          | List | Contains jobs of queue {QUEUE} ordered by insertion time
`{NS}:queue:{QUEUE}:failed`   | List | Contains failed jobs of queue {QUEUE} ordered by failing time
`{NS}:queue:{QUEUE}:stats`   | Hash | Contains stats about a queue
`{NS}:workers`                | Set  | Contains names of active workers
`{NS}:worker:{HOSTNAME}:{ID}` | Hash | Contains stats and state about a worker


### `{NS}:queues`

A basic Redis Set containing names of every active queues.
When a job is enqueued to a queue, {QUEUE} name is automatically added to this Set.
There is currently no way to removes it except manually

```
['emails', 'batch', 'login_events']
```

### `{NS}:queue:{QUEUE}`

This list contains every pending jobs for a queue.

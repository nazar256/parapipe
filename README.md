Parapipe - paralleling pipeline (UNDER DEVELOPMENT)
===============================

The library provides a buffered pipeline for structuring the code and vertically scaling your app. The main difference
from a regular pipeline example you may find on the internet - pipeline executes everything on each step concurrently,
yet maintaining the order. Although, this library does not use any locks or mutexes, or any other thread synchronization
tools. Just channels and nothing more.

When to use
-----------

Installation
------------

```
go get github.com/nazar256/parapipe
```

Usage
-----

It's best shown in examples.
Every box executed in each pipe in parallel.

### AMQP middleware

```go
	repliesCh, err = amqpChannel.Consume(
		q.Name,            // queue
		"go-amqp-example", // consumer
		true,              // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)

	bufferSize := 4
	pipeline := parapipe.
		NewPipeline(repliesCh, bufferSize).
		// parse
		Pipe(func(box parapipe.Box) parapipe.Box {
			msg := box.Contents
			event := &Event{}
			_ = json.Unmarshal(msg.(amqp.Delivery).Body, event)
			return parapipe.Box{event}
		}).
		// validate
		Pipe(func(box parapipe.Box) parapipe.Box {
			event := box.Contents.(Event)
			// ...
			return parapipe.Box{event}
		}).
		// modify
		Pipe(func(box parapipe.Box) parapipe.Box {
			event := box.Contents.(Event)
			// ...
            payload, _ := json.Marshal(event)
			return parapipe.Box{payload}
		}).
		// publish as new event
		Pipe(func(box parapipe.Box) parapipe.Box {
			payload, _ := box.Contents.([]byte)
		    err = amqpChannel.Publish(
			"some-exchange",
			"some:routing:key",
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        payload,
			})
			// validate message
			return parapipe.Box{nil}
		})
```

### Streamed feed response
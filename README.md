Parapipe - paralleling pipeline
===============================

The library provides a zero-dependency non-blocking buffered FIFO-pipeline for structuring the code and vertically scaling your app. 
The main difference from a regular pipeline example you may find on the internet - pipeline executes everything on each step concurrently,
yet maintaining the order. Although, this library does not use any locks or mutexes, or any other thread synchronization
tools. Just pure channels.

When to use
-----------

* processed data can be divided in chunks (messages), and the flow may consist of one or more stages
* data should be processed concurrently (scaled vertically)
* the order of processing messages must be maintained

Installation
------------

```
go get github.com/nazar256/parapipe
```

Usage
-----

1. Create a pipeline

```go
concurrencyFactor := 5 // how many callbacks can be executed concurrently for each pipe
pipeline := parapipe.NewPipeline(concurrencyFactor)
```

2. Add pipes - call `Pipe()` method one or more times
```go
pipeline.Pipe(func(msg interface{}) interface{} {
    typedMsg := msg.(YourInputType)     // assert your type for the message
    // do something and generate a new value "someValue"
    return someValue
})
   ```
3. Get "out" channel when all pipes are added and read results from it
```go
for result := range pipeline.Out() {
    // do something with the result
}
```
4. Close pipeline with closing it's input channel. All internal channels, goroutines, including `Out()` channel will be
   closed in a cascade.
```go
close(pipeline.In())
```   

### Limitations

* `Out()` method can be used only once on each pipeline. Any subsequent `Pipe()` call will cause panic. Though, when you
  need to stream values somewhere from the middle of the pipeline - just send them to your own channel.
* as at the time of writing Go does not have generics, you have to assert the type for incoming messages in pipes explicitly.

Examples
--------

### AMQP middleware

Parapipe can be handful when you need to process messages in the middle concurrently, yet maintaining their order.

```go
replies, err = amqpChannel.Consume(
    q.Name,            // queue
    "go-amqp-example", // consumer
    true,              // auto-ack
    false,             // exclusive
    false,             // no-local
    false,             // no-wait
    nil,               // args
)

concurrency := 4
pipeSource := interface{}(replies).(<-chan interface{})
pipeline := parapipe.NewPipeline(concurrency)

go func() {
    for amqpMsg := range replies {
        pipeline.In() <- amqpMsg
    }
    close(pipeline.In())
}()

// here pipeline starts to process messages immediately (once they are sent above) even before "Out()" is called
pipeline.
    Pipe(func(msg interface{}) interface{} {
        event := &Event{}
        _ = json.Unmarshal(msg.(amqp.Delivery).Body, event)
        return event
    }).
    Pipe(func(msg interface{}) interface{} {
        event := msg.(Event)
        // validate the event
        return event
    }).
    Pipe(func(msg interface{}) interface{} {
        payload, _ := json.Marshal(msg.(Event))
        // publish as new event
        err = amqpChannel.Publish(
        "some-exchange",
        "some:routing:key",
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        payload,
        })
        return nil
    })
```

### Other examples

With parapipe you can:

  * respond a JSON-feed as stream, retrieve, enrich and marshal each object concurrently, in maintained order and return them to the client
  * fetch and merge entries from different sources as one stream
  * structure your HTTP-controllers
  * processing heavy files in effective way

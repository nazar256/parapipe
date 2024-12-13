Parapipe - paralleling pipeline
===============================

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go)
[![tests](https://github.com/nazar256/parapipe/actions/workflows/tests.yml/badge.svg)](https://github.com/nazar256/parapipe/actions/workflows/tests.yml)
[![linters](https://github.com/nazar256/parapipe/actions/workflows/linters.yml/badge.svg)](https://github.com/nazar256/parapipe/actions/workflows/linters.yml)
[![coverage](https://codecov.io/gh/nazar256/parapipe/branch/main/graph/badge.svg?token=N6NI66KPXG)](https://codecov.io/gh/nazar256/parapipe)
[![Go Report Card](https://goreportcard.com/badge/github.com/nazar256/parapipe)](https://goreportcard.com/report/github.com/nazar256/parapipe)
[![GoDoc](https://godoc.org/github.com/nazar256/parapipe?status.svg)](https://godoc.org/github.com/nazar256/parapipe)

The library provides a zero-dependency non-blocking buffered FIFO-pipeline 
for structuring the code and vertically scaling your app. 
Unlike regular pipeline examples you may find on the internet - parapipe executes everything on each step concurrently,
yet maintaining the output order. Although, this library does not use any locks or mutexes. Just pure channels.

When to use
-----------

* processed data can be divided in chunks (messages), and the flow may consist of one or more stages
* data should be processed concurrently (scaled vertically)
* the order of processing messages must be maintained

Installation
------------

```
go get -u github.com/nazar256/parapipe@latest
```

Usage
-----

1. Create a pipeline with first step. Processing callback is generic (so as the pipeline). 
It may receive and return any type of data, but the second return value should always be a boolean.

```go
concurrency := runtime.NumCPU()     // how many messages to process concurrently for each pipe
pipeline := parapipe.NewPipeline(concurrency, func(msg YourInputType) (YourOutputType, bool) {
    // do something and generate a new value "someValue"
    shouldProceedWithNextStep := true
    return someValue, shouldProceedWithNextStep
})
```

2. Add pipes - call `Attach()` function one or more times to add steps to the pipeline
```go
p1 := parapipe.NewPipeline(runtime.NumCPU(), func(msg int) (int, bool) {
    time.Sleep(30 * time.Millisecond)
    return msg + 1000, true
})
p2 := parapipe.Attach(p1, parapipe.NewPipeline(concurrency, func(msg int) (string, bool) {
    time.Sleep(30 * time.Millisecond)
    return strconv.Itoa(msg), true
}))

// final pipeline you are going to work with (push messages and read output)
pipeline := parapipe.Attach(p2, parapipe.NewPipeline(concurrency, func(msg string) (string, bool) {
    time.Sleep(30 * time.Millisecond)
    return "#" + msg, true
}))
```

3. Get "out" channel when all pipes are added and read results from it
```go
for result := range pipeline.Out() {
    // do something with the result
}
```
It's **important** to drain the pipeline (read everything from "out") even when the pipeline won't produce any viable result. 
It could be stuck otherwise.

4. Push values for processing into the pipeline:
```go
pipeline.Push("something")
```   

5. Close pipeline to after the last message. This will cleanup its resources and close its output channel. 
   It's not recommended closing pipeline using `defer` because you may not want to hang output util defer is executed.
```go
pipeline.Close()
```   

### Circuit breaking

In some cases (errors) there could be impossible to process a message, thus there is no way to pass it further.
In such case just return `false` as a second return value from the step processing callback. 
The first value will be ignored.

```go
pipeline.Pipe(4, func(inputValue InputType) (OutputType, bool) {
    someValue, err := someOperation(inputValue)
    if err != nil {
		// handle the error
		// slog.Error("error when calling someOperation", "err", err)
        return someValue, false
    }
    return someValue, true
})
// ...
for result := range pipeline.Out() {
    // do something with the result
}
```

### Performance

Parapipe makes use of generics and channels.
Overall it should be performant enough for most of the cases.
It has zero heap allocations in hot code, thus generates little load for garbage collector.
However, it uses channels under the hood and is bottlenecked mostly by the channel operations which are several
writes and reads per each message.

Examples
--------

### AMQP middleware

Parapipe can be handful when you need to process messages in the middle concurrently, yet maintaining their order.

See the [working example of using parapipe in AMQP client](http://github.com/nazar256/go-amqp-sniffer/blob/a5c5db375dc68a2e83c24686e4e57a63cf08c80b/sniffer/sniffer.go#L49-L108).

### Other examples

With parapipe you can:

  * in your API respond a long JSON-feed as stream, retrieve, enrich and marshal each object concurrently, in maintained order and return them to the client
  * fetch and merge entries from different sources as one stream
  * structure your API controllers or handlers
  * processing heavy files in effective way

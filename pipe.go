package parapipe

import (
	"log/slog"
	"sync"
)

type Callback[I any, O any] func(msg I) (output O, proceed bool)

type pipe[I any, O any] struct {
	in          chan I
	out         chan O
	callback    Callback[I, O]
	concurrency int
}

// newPipe creates a new pipe with the specified concurrency and callback function.
func newPipe[I, O any](concurrency int, callback Callback[I, O]) *pipe[I, O] {
	if concurrency < 1 {
		slog.Error("concurrency must be greater than 0", slog.Int("concurrency", concurrency))
		concurrency = 1
	}

	p := pipe[I, O]{
		in:          make(chan I, 1),
		out:         make(chan O, concurrency+1),
		callback:    callback,
		concurrency: concurrency,
	}

	promisesCh := make(chan promise[O], concurrency)

	wp := startWorkerPool[I, O](callback, concurrency, newPromisePool[O]())

	// queue message processing and output with promises
	go func() {
		for msg := range p.in {
			promisesCh <- wp.push(msg)
		}

		close(promisesCh)
	}()

	// wait for each promise and close outCh
	go func() {
		for pr := range promisesCh {
			value, ok := pr.await()
			if !ok {
				continue
			}

			p.out <- value
		}

		close(p.out)
	}()

	return &p
}

type promiseValue[O any] struct {
	value O
	ok    bool
}

type promise[O any] struct {
	pool *sync.Pool
	ch   chan promiseValue[O]
}

type promisePool[O any] struct {
	pool *sync.Pool
}

func newPromisePool[O any]() promisePool[O] {
	return promisePool[O]{
		pool: &sync.Pool{
			New: func() any {
				return make(chan promiseValue[O], 1)
			},
		},
	}
}

func (pp promisePool[O]) get() promise[O] {
	return promise[O]{
		pool: pp.pool,
		ch:   pp.pool.Get().(chan promiseValue[O]),
	}
}

func (p promise[O]) send(v O) {
	p.ch <- promiseValue[O]{
		value: v,
		ok:    true,
	}
}

func (p promise[O]) cancel() {
	p.ch <- promiseValue[O]{
		ok: false,
	}
}

func (p promise[O]) await() (O, bool) {
	result := <-p.ch
	p.pool.Put(p.ch)

	return result.value, result.ok
}

type asyncJob[I, O any] struct {
	msg     I
	promise promise[O]
}

type workerPool[I, O any] struct {
	queue       chan asyncJob[I, O]
	promisePool promisePool[O]
}

func startWorkerPool[I, O any](job Callback[I, O], concurrency int, promisePool promisePool[O]) *workerPool[I, O] {
	wp := &workerPool[I, O]{
		queue:       make(chan asyncJob[I, O], concurrency+1),
		promisePool: promisePool,
	}

	for i := 0; i < concurrency; i++ {
		go func() {
			for j := range wp.queue {
				result, proceed := job(j.msg)
				if !proceed {
					j.promise.cancel()
					continue
				}

				j.promise.send(result)
			}
		}()
	}

	return wp
}

func (wp *workerPool[I, O]) push(msg I) promise[O] {
	p := wp.promisePool.get()

	wp.queue <- asyncJob[I, O]{
		msg:     msg,
		promise: p,
	}

	return p
}

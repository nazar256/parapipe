package parapipe

import "log/slog"

type Config struct{}

type Pipeline[I, O any] struct {
	inCh   chan<- I
	outCh  <-chan O
	closed bool
}

func NewPipeline[I, O any](concurrency int, callback Callback[I, O]) *Pipeline[I, O] {
	p := newPipe[I, O](concurrency, callback)

	return &Pipeline[I, O]{
		inCh:  p.in,
		outCh: p.out,
	}
}

func (p *Pipeline[I, O]) Push(v I) {
	if p.closed {
		slog.Error("cannot push after Close is called")
		return
	}
	p.inCh <- v
}

func (p *Pipeline[I, O]) Out() <-chan O {
	return p.outCh
}

func (p *Pipeline[I, O]) Close() {
	if p.closed {
		return
	}

	p.closed = true

	close(p.inCh)
}

func Attach[I, T, O any](left *Pipeline[I, T], right *Pipeline[T, O]) *Pipeline[I, O] {
	go func() {
		for v := range left.Out() {
			right.Push(v)
		}

		right.Close()
	}()

	return &Pipeline[I, O]{
		inCh:   left.inCh,
		outCh:  right.outCh,
		closed: left.closed || right.closed,
	}
}

package parapipe

import "runtime"

// Pipeline executes jobs concurrently maintaining message order
type Pipeline struct {
	cfg    Config
	pipes  []*pipe
	hasOut bool
	closed bool
}

// Config contains pipeline parameters which influence execution or behavior
type Config struct {
	Concurrency   int  // how many messages to process concurrently for each pipe
	ProcessErrors bool // if false, messages implementing "error" interface will not be passed to subsequent workers
}

// NewPipeline creates new pipeline instance, "Concurrency" sets how many jobs can be executed concurrently in each pipe
func NewPipeline(cfg Config) *Pipeline {
	if cfg.Concurrency < 1 {
		cfg.Concurrency = runtime.NumCPU()
	}

	return &Pipeline{
		pipes: make([]*pipe, 0, 1),
		cfg:   cfg,
	}
}

// Push adds a value to the pipeline for processing, it is immediately queued to be processed
func (p *Pipeline) Push(v interface{}) {
	p.pipes[0].in <- v
}

// Pipe adds new pipe to pipeline with the callback for processing each message
func (p *Pipeline) Pipe(job Job) *Pipeline {
	if p.hasOut || p.closed {
		panic("attempt to create new pipeline after Out() call")
	}

	pipe := newPipe(job, p.cfg.Concurrency, p.cfg.ProcessErrors)

	if len(p.pipes) > 0 {
		bindChannels(p.pipes[len(p.pipes)-1].out, pipe.in)
	}

	p.pipes = append(p.pipes, pipe)

	return p
}

// Out returns exit of the pipeline - channel with results of the last pipe. Call it once - it is not idempotent!
func (p *Pipeline) Out() <-chan interface{} {
	p.hasOut = true
	return p.pipes[len(p.pipes)-1].out
}

// Close closes pipeline input channel, from that moment pipeline processes what is left and releases the resources
// it must not be used after Close is called
func (p *Pipeline) Close() {
	p.closed = true
	close(p.pipes[0].in)
}

func bindChannels(from <-chan interface{}, to chan<- interface{}) {
	go func(from <-chan interface{}, to chan<- interface{}) {
		for msg := range from {
			to <- msg
		}

		close(to)
	}(from, to)
}

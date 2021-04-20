package parapipe

// Pipeline executes jobs concurrently maintaining message order
type Pipeline struct {
	cfg       Config
	pipes     []*pipe
	finalized bool
}

// Config contains pipeline parameters which influence execution or behavior
type Config struct {
	concurrency   int
	processErrors bool
}

// NewPipeline creates new Pipeline instance, "concurrency" sets how many jobs can be executed concurrently in each pipe
func NewPipeline(cfg Config) *Pipeline {
	return &Pipeline{
		pipes: make([]*pipe, 0, 1),
		cfg:   cfg,
	}
}

// In returns input channel of the pipeline - the entrance of the pipeline, send there your input values
func (p *Pipeline) In() chan<- interface{} {
	return p.pipes[0].in
}

// Pipe adds new pipe to pipeline with the callback for processing each message
func (p *Pipeline) Pipe(job Job) *Pipeline {
	if p.finalized {
		panic("attempt to create new pipeline after Out() call")
	}

	pipe := newPipe(job, p.cfg.concurrency, p.cfg.processErrors)

	if len(p.pipes) > 0 {
		bindChannels(p.pipes[len(p.pipes)-1].out, pipe.in)
	}

	p.pipes = append(p.pipes, pipe)

	return p
}

// Out returns exit of the pipeline - channel with results of the last pipe. Call it once - it is not idempotent!
func (p *Pipeline) Out() <-chan interface{} {
	p.finalized = true
	return p.pipes[len(p.pipes)-1].out
}

func bindChannels(from <-chan interface{}, to chan<- interface{}) {
	go func(from <-chan interface{}, to chan<- interface{}) {
		for msg := range from {
			to <- msg
		}
		close(to)
	}(from, to)
}

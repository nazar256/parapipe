package parapipe

type pipe struct {
	in            chan interface{}
	out           chan interface{}
	processErrors bool

	queue     chan chan interface{}
	closeInCh chan struct{}
}

// Job is a short callback signature, used in pipes
type Job func(msg interface{}) interface{}

func newPipe(job Job, concurrency int, processErrors bool) *pipe {
	p := &pipe{
		in:            make(chan interface{}, 1),
		out:           make(chan interface{}, 1),
		processErrors: processErrors,

		queue:     make(chan chan interface{}, concurrency),
		closeInCh: make(chan struct{}),
	}

	go func() {
		for msg := range p.in {
			queued := make(chan interface{}, 1)

			_, isError := msg.(error)
			if isError && !p.processErrors {
				queued <- msg
			} else {
				go func(job Job, msg interface{}, queued chan interface{}) {
					queued <- job(msg)
				}(job, msg, queued)
			}
			p.queue <- queued
		}
		close(p.queue)
	}()

	go func() {
		for processed := range p.queue {
			p.out <- <-processed
			close(processed)
		}

		close(p.out)
	}()

	return p
}

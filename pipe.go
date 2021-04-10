package parapipe

type pipe struct {
	in  chan interface{}
	out chan interface{}

	queue     chan chan interface{}
	closeInCh chan struct{}
}

// Job is the short callback signature, used in pipes
type Job func(msg interface{}) interface{}

func newPipe(job Job, concurrency int) *pipe {
	p := &pipe{
		in:  make(chan interface{}),
		out: make(chan interface{}),

		queue:     make(chan chan interface{}, concurrency),
		closeInCh: make(chan struct{}),
	}

	go func() {
		for msg := range p.in {
			queued := make(chan interface{}, 1)
			go func(job Job, msg interface{}, queued chan interface{}) {
				queued <- job(msg)
			}(job, msg, queued)
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

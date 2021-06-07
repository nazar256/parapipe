package parapipe

type pipe struct {
	in            jobChan
	out           jobChan
	processErrors bool

	queue     chan queuedJobChan
	closeInCh chan struct{}
}

// Job is a short callback signature, used in pipes
type Job func(msg interface{}) interface{}

type jobChan chan interface{}
type queuedJobChan chan interface{}

func newPipe(job Job, concurrency int, processErrors bool) *pipe {
	p := &pipe{
		in:            make(jobChan, 1),
		out:           make(jobChan, 1),
		processErrors: processErrors,

		queue:     make(chan queuedJobChan, concurrency),
		closeInCh: make(chan struct{}),
	}

	go func() {
		for msg := range p.in {
			queued := make(queuedJobChan, 1)

			_, isError := msg.(error)
			if isError && !p.processErrors {
				queued <- msg
			} else {
				go func(job Job, msg interface{}, queued queuedJobChan) {
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

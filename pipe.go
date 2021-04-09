package parapipe

type Box struct {
	Contents interface{}
}

type pipe struct {
	in      chan Box
	out     chan Box
	job     func(inBox Box) Box
	queue   chan chan Box
	closeCh chan struct{}
}

func newPipe(job func(inBox Box) Box, buffer int) *pipe {
	p := &pipe{
		in:  make(chan Box, buffer),
		out: make(chan Box, buffer),
		job: job,

		queue:   make(chan chan Box, buffer),
		closeCh: make(chan struct{}),
	}

	go func() {
		<-p.closeCh
		close(p.in)
		close(p.queue)
		close(p.out)
		return
	}()

	go func() {
		for box := range p.in {
			queued := make(chan Box, 1)
			go func(box Box) {
				queued <- p.job(box)
			}(box)
			p.queue <- queued
		}
	}()

	go func() {
		for processed := range p.queue {
			p.out <- <-processed
			close(processed)
		}
	}()

	return p
}

func (p pipe) close() {
	p.closeCh <- struct{}{}
}

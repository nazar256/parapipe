package parapipe

type Pipeline struct {
	source <-chan interface{}
	pipes  []*pipe
	buffer int
}

func NewPipeline(source <-chan interface{}, buffer int) *Pipeline {
	return &Pipeline{
		source: source,
		pipes:  make([]*pipe, 0),
		buffer: buffer,
	}
}

func (p Pipeline) Pipe(job func(box Box) Box) Pipeline {
	pipe := newPipe(job, p.buffer)

	p.pipes = append(p.pipes, pipe)

	var source <-chan Box
	switch len(p.pipes) {
	case 0:
		source = packCh(p.source)
	default:
		source = p.pipes[len(p.pipes)-1].out
	}

	go func() {
		for box := range source {
			pipe.in <- box
		}
	}()

	return p
}

func (p Pipeline) Out() <-chan interface{} {
	return unPackCh(p.pipes[len(p.pipes)-1].out)
}

func (p Pipeline) Close() {
	for _, pipe := range p.pipes {
		pipe.close()
	}
}

func packCh(inCh <-chan interface{}) chan Box {
	outCh := make(chan Box, cap(inCh))
	go func() {
		for msg := range inCh {
			outCh <- Box{msg}
		}
		close(outCh)
	}()
	return outCh
}

func unPackCh(inCh chan Box) chan interface{} {
	outCh := make(chan interface{}, cap(inCh))
	go func() {
		for box := range inCh {
			outCh <- box.Contents
		}
		close(outCh)
	}()
	return outCh
}

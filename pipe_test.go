package parapipe

import (
	"math/rand"
	"testing"
	"testing/quick"
	"time"
)

func genIntMessages(dst chan<- interface{}, amount int) {
	go func() {
		for i := 0; i < amount; i++ {
			dst <- i
		}
		close(dst)
	}()
}

func TestPipeMaintainsMessageOrder(t *testing.T) {
	pipe := newPipe(func(msg interface{}) interface{} {
		return msg.(int) + 1
	}, rand.Intn(4), false)

	msgAmount := 100 + rand.Intn(1000)

	go func() {
		for input := 0; input < msgAmount; input++ {
			pipe.in <- input
		}

		close(pipe.in)
	}()

	for input := 0; input < msgAmount; input++ {
		expected := input + 1
		actual := <-pipe.out

		if actual != expected {
			t.Errorf("order not maintained, expected %d in sequence of %d, got %d", expected, msgAmount, actual)
			t.Fail()
		}
	}
}

func TestPipeExecutesJobsConcurrently(t *testing.T) {
	concurrency := 100

	pipe := newPipe(func(msg interface{}) interface{} {
		time.Sleep(time.Duration(concurrency) * time.Millisecond)
		return msg
	}, concurrency, false)

	start := time.Now()

	genIntMessages(pipe.in, concurrency)

	// wait for everything is processed
	for range pipe.out {
	}

	duration := time.Since(start)
	if duration > 150*time.Millisecond {
		t.Errorf(
			"Expected to be executed concurrently in 100ms, actual duration was %dms",
			duration/time.Millisecond,
		)
	}
}

func TestPipeCanSkipErrorProcessing(t *testing.T) {
	concurrency := rand.Intn(20)
	pipe := newPipe(func(msg interface{}) interface{} {
		err := msg.(error)
		if err != nil {
			t.Error("error expected to not be passed to worker, but it was")
		}
		return msg
	}, concurrency, false)

	pipe.in <- new(quick.CheckError)
	close(pipe.in)
	<-pipe.out
}

func TestPipeCanProcessErrors(t *testing.T) {
	concurrency := rand.Intn(20)
	errProcessed := false
	pipe := newPipe(func(msg interface{}) interface{} {
		_, isErr := msg.(error)
		if isErr {
			errProcessed = true
		}
		return msg
	}, concurrency, true)

	pipe.in <- new(quick.CheckError)
	close(pipe.in)
	<-pipe.out

	if !errProcessed {
		t.Error("error expected to be passed to worker, but it was not")
	}
}

package parapipe

import (
	"math/rand"
	"testing"
	"time"
)

func genIntMessages(dst chan<- int, amount int) {
	go func() {
		for i := 0; i < amount; i++ {
			dst <- i
		}

		close(dst)
	}()
}

func TestPipeMaintainsMessageOrder(t *testing.T) {
	p := newPipe[int, int](rand.Intn(4), func(msg int) (int, bool) {
		return msg + 1, true
	})

	msgAmount := 100 + rand.Intn(1000)

	go func() {
		for input := 0; input < msgAmount; input++ {
			p.in <- input
		}

		close(p.in)
	}()

	for input := 0; input < msgAmount; input++ {
		expected := input + 1
		actual := <-p.out

		if actual != expected {
			t.Errorf("order not maintained, expected %d in sequence of %d, got %d", expected, msgAmount, actual)
			t.Fail()
		}
	}
}

func TestPipeExecutesJobsConcurrently(t *testing.T) {
	concurrency := 100

	p := newPipe[int, int](concurrency, func(msg int) (int, bool) {
		time.Sleep(time.Duration(concurrency) * time.Millisecond)
		return msg, true
	})

	start := time.Now()

	genIntMessages(p.in, concurrency)

	// wait for everything is processed
	for range p.out {
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
	p := newPipe[int, int](rand.Intn(20), func(msg int) (int, bool) {
		return msg + 1, false
	})

	go func() {
		p.in <- 1
		close(p.in)
	}()

	outValue := <-p.out

	if outValue != 0 {
		t.Error("error processing should skip the job")
	}
}

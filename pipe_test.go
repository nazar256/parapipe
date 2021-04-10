package parapipe

import (
	"math/rand"
	"testing"
	"time"
)

func TestPipeMaintainsMessageOrder(t *testing.T) {
	pipe := newPipe(func(msg interface{}) interface{} {
		return msg.(int) + 1
	}, rand.Intn(4))

	inputValues := makeRange(100, 100+rand.Intn(1000))

	go func() {
		for input := range inputValues {
			pipe.in <- input
		}
		close(pipe.in)
	}()

	for input := range inputValues {
		expected := input + 1
		actual := <-pipe.out

		if actual != expected {
			t.Errorf("order not maintained, expected %d in sequence of %d, got %d", expected, len(inputValues), actual)
			t.Fail()
		}
	}
}

func TestPipeExecutesJobsConcurrently(t *testing.T) {
	concurrency := 100

	pipe := newPipe(func(msg interface{}) interface{} {
		time.Sleep(time.Duration(concurrency) * time.Millisecond)
		return msg
	}, concurrency)

	start := time.Now()

	genMessages(pipe.in, concurrency)

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

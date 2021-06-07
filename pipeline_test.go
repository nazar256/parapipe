package parapipe_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"testing/quick"
	"time"

	"github.com/nazar256/parapipe"
)

func TestPipelineExecutesPipesInDefinedOrder(t *testing.T) {
	concurrency := rand.Intn(5) + 2
	pipeline := parapipe.NewPipeline(parapipe.Config{}).
		Pipe(concurrency, func(msg interface{}) interface{} {
			number := msg.(int)
			return strconv.Itoa(number)
		}).
		Pipe(concurrency, func(msg interface{}) interface{} {
			str := msg.(string)
			return "#" + str
		})

	feedPipeline(pipeline, 100)

	i := 0

	for actualResult := range pipeline.Out() {
		expected := fmt.Sprintf("#%d", i)
		if actualResult != expected {
			t.Errorf("got wrong result from pipeline at iteration %d: \"%s\" instead of \"%s\"", i, actualResult, expected)
			t.Fail()
		}
		i++
	}
}

func TestPipelineExecutesConcurrently(t *testing.T) {
	inputValuesCount := 100
	concurrency := inputValuesCount
	pipeline := parapipe.NewPipeline(parapipe.Config{}).
		Pipe(concurrency, func(msg interface{}) interface{} {
			time.Sleep(30 * time.Millisecond)
			return msg.(int) + 1000
		}).
		Pipe(concurrency, func(msg interface{}) interface{} {
			time.Sleep(30 * time.Millisecond)
			return strconv.Itoa(msg.(int))
		}).
		Pipe(concurrency, func(msg interface{}) interface{} {
			time.Sleep(30 * time.Millisecond)
			return "#" + msg.(string)
		})

	start := time.Now()

	feedPipeline(pipeline, inputValuesCount)

	// wait for all results
	for range pipeline.Out() {
	}

	if time.Since(start) > 150*time.Millisecond {
		t.Errorf(
			"Expected to be executed concurrently in 100ms, actual duration was %dms",
			time.Since(start)/time.Millisecond,
		)
	}
}

func TestPipelineSkipsErrorsByDefault(t *testing.T) {
	concurrency := rand.Intn(20)
	pipeline := parapipe.NewPipeline(parapipe.Config{}).
		Pipe(concurrency, func(msg interface{}) interface{} {
			_, isErr := msg.(error)
			if isErr {
				t.Error("Error expected to be skipped from processing, but worker has received one")
			}
			return msg
		}).
		Pipe(concurrency, func(msg interface{}) interface{} {
			_, isErr := msg.(error)
			if isErr {
				t.Error("Error expected to be skipped from processing, but worker has received one")
			}
			return msg
		})

	pipeline.Push(new(quick.CheckError))
	pipeline.Close()
	<-pipeline.Out()
}

func TestPipelineCanProcessErrors(t *testing.T) {
	errorProcessCount := 0
	pipeline := parapipe.NewPipeline(parapipe.Config{ProcessErrors: true}).
		Pipe(0, func(msg interface{}) interface{} {
			_, isErr := msg.(error)
			if isErr {
				errorProcessCount++
			}
			return msg
		}).
		Pipe(0, func(msg interface{}) interface{} {
			_, isErr := msg.(error)
			if isErr {
				errorProcessCount++
			}
			return msg
		})

	pipeline.Push(new(quick.CheckError))
	pipeline.Close()
	<-pipeline.Out()

	if errorProcessCount != 2 {
		t.Error("error expected to be passed to both workers, but it was not")
	}
}

func Benchmark1Pipe1Message(b *testing.B) {
	for n := 0; n < b.N; n++ {
		concurrency := 1
		pipeline := parapipe.NewPipeline(parapipe.Config{}).
			Pipe(concurrency, func(msg interface{}) interface{} { return msg })
		feedPipeline(pipeline, 1)

		for range pipeline.Out() {
		}
	}
}

func Benchmark5Pipes1Message(b *testing.B) {
	for n := 0; n < b.N; n++ {
		concurrency := 1
		pipeline := parapipe.NewPipeline(parapipe.Config{}).
			Pipe(concurrency, func(msg interface{}) interface{} { return msg }).
			Pipe(concurrency, func(msg interface{}) interface{} { return msg }).
			Pipe(concurrency, func(msg interface{}) interface{} { return msg }).
			Pipe(concurrency, func(msg interface{}) interface{} { return msg }).
			Pipe(concurrency, func(msg interface{}) interface{} { return msg })
		feedPipeline(pipeline, 1)

		for range pipeline.Out() {
		}
	}
}

func Benchmark1Pipe10000Messages(b *testing.B) {
	concurrency := 8
	pipeline := parapipe.NewPipeline(parapipe.Config{}).
		Pipe(concurrency, func(msg interface{}) interface{} { return msg })
	msgCount := 10000

	for n := 0; n < b.N; n++ {
		go func() {
			for i := 0; i < msgCount; i++ {
				pipeline.Push(i)
			}
		}()

		for i := 0; i < msgCount; i++ {
			<-pipeline.Out()
		}
	}

	pipeline.Close()
}

func Benchmark1Pipe10000MessagesBatchedBy100(b *testing.B) {
	concurrency := 8
	pipeline := parapipe.NewPipeline(parapipe.Config{}).
		Pipe(concurrency, func(msg interface{}) interface{} { return msg })
	batchCount := 100
	batchSize := 100

	for n := 0; n < b.N; n++ {
		go func() {
			msg := makeRange(1, batchSize)

			for i := 0; i < batchCount; i++ {
				pipeline.Push(msg)
			}
		}()

		for i := 0; i < batchCount; i++ {
			<-pipeline.Out()
		}
	}

	pipeline.Close()
}

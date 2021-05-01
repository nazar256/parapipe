package parapipe

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"testing/quick"
	"time"
)

func TestPipelineExecutesPipesInDefinedOrder(t *testing.T) {
	pipeline := NewPipeline(Config{Concurrency: rand.Intn(5) + 2}).
		Pipe(func(msg interface{}) interface{} {
			number := msg.(int)
			return strconv.Itoa(number)
		}).
		Pipe(func(msg interface{}) interface{} {
			str := msg.(string)
			return "#" + str
		})

	genIntMessages(pipeline.In(), 100)

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
	pipeline := NewPipeline(Config{Concurrency: inputValuesCount}).
		Pipe(func(msg interface{}) interface{} {
			time.Sleep(30 * time.Millisecond)
			return msg.(int) + 1000
		}).
		Pipe(func(msg interface{}) interface{} {
			time.Sleep(30 * time.Millisecond)
			return strconv.Itoa(msg.(int))
		}).
		Pipe(func(msg interface{}) interface{} {
			time.Sleep(30 * time.Millisecond)
			return "#" + msg.(string)
		})

	start := time.Now()

	genIntMessages(pipeline.In(), inputValuesCount)

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
	cnc := rand.Intn(20)
	pipeline := NewPipeline(Config{Concurrency: cnc}).
		Pipe(func(msg interface{}) interface{} {
			_, isErr := msg.(error)
			if isErr {
				t.Error("Error expected to be skipped from processing, but worker has received one")
			}
			return msg
		}).
		Pipe(func(msg interface{}) interface{} {
			_, isErr := msg.(error)
			if isErr {
				t.Error("Error expected to be skipped from processing, but worker has received one")
			}
			return msg
		})

	pipeline.In() <- new(quick.CheckError)
	close(pipeline.In())
	<-pipeline.Out()
}

func TestPipelineCanProcessErrors(t *testing.T) {
	cnc := rand.Intn(20)
	errorProcessCount := 0
	pipeline := NewPipeline(Config{Concurrency: cnc, ProcessErrors: true}).
		Pipe(func(msg interface{}) interface{} {
			_, isErr := msg.(error)
			if isErr {
				errorProcessCount++
			}
			return msg
		}).
		Pipe(func(msg interface{}) interface{} {
			_, isErr := msg.(error)
			if isErr {
				errorProcessCount++
			}
			return msg
		})

	pipeline.In() <- new(quick.CheckError)
	close(pipeline.In())
	<-pipeline.Out()

	if errorProcessCount != 2 {
		t.Error("error expected to be passed to both workers, but it was not")
	}
}

func Benchmark1Pipe1Message(b *testing.B) {
	for n := 0; n < b.N; n++ {
		concurrency := 1
		pipeline := NewPipeline(Config{Concurrency: concurrency}).Pipe(func(msg interface{}) interface{} { return msg })
		genIntMessages(pipeline.In(), 1)

		for range pipeline.Out() {
		}
	}
}

func Benchmark5Pipes1Message(b *testing.B) {
	for n := 0; n < b.N; n++ {
		concurrency := 1
		pipeline := NewPipeline(Config{Concurrency: concurrency}).
			Pipe(func(msg interface{}) interface{} { return msg }).
			Pipe(func(msg interface{}) interface{} { return msg }).
			Pipe(func(msg interface{}) interface{} { return msg }).
			Pipe(func(msg interface{}) interface{} { return msg }).
			Pipe(func(msg interface{}) interface{} { return msg })
		genIntMessages(pipeline.In(), 1)

		for range pipeline.Out() {
		}
	}
}

func Benchmark1Pipe10000Messages(b *testing.B) {
	concurrency := 8
	pipeline := NewPipeline(Config{Concurrency: concurrency}).Pipe(func(msg interface{}) interface{} { return msg })
	msgCount := 10000

	for n := 0; n < b.N; n++ {
		go func() {
			for i := 0; i < msgCount; i++ {
				pipeline.In() <- i
			}
		}()

		for i := 0; i < msgCount; i++ {
			<-pipeline.Out()
		}
	}

	close(pipeline.In())
}

func Benchmark1Pipe10000MessagesBatchedBy100(b *testing.B) {
	concurrency := 8
	pipeline := NewPipeline(Config{Concurrency: concurrency}).Pipe(func(msg interface{}) interface{} { return msg })
	batchCount := 100
	batchSize := 100

	for n := 0; n < b.N; n++ {
		go func() {
			msg := makeRange(1, batchSize)

			for i := 0; i < batchCount; i++ {
				pipeline.In() <- msg
			}
		}()

		for i := 0; i < batchCount; i++ {
			<-pipeline.Out()
		}
	}

	close(pipeline.In())
}

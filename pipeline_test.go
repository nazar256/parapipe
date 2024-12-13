package parapipe_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/nazar256/parapipe"
)

func TestPipelineExecutesPipesInDefinedOrder(t *testing.T) {
	concurrency := rand.Intn(5) + 2
	p1 := parapipe.NewPipeline(concurrency, func(msg int) (string, bool) {
		return strconv.Itoa(msg), true
	})
	pipeline := parapipe.Attach(p1, parapipe.NewPipeline(concurrency, func(msg string) (string, bool) {
		return "#" + msg, true
	}))

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
	p1 := parapipe.NewPipeline(concurrency, func(msg int) (int, bool) {
		time.Sleep(30 * time.Millisecond)
		return msg + 1000, true
	})
	p2 := parapipe.Attach(p1, parapipe.NewPipeline(concurrency, func(msg int) (string, bool) {
		time.Sleep(30 * time.Millisecond)
		return strconv.Itoa(msg), true
	}))
	pipeline := parapipe.Attach(p2, parapipe.NewPipeline(concurrency, func(msg string) (string, bool) {
		time.Sleep(30 * time.Millisecond)
		return "#" + msg, true
	}))

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

func TestPipelineSkipsMessages(t *testing.T) {
	concurrency := rand.Intn(20)
	p1 := parapipe.NewPipeline(concurrency, func(msg int) (int, bool) {
		return msg + 1, false
	})
	pipeline := parapipe.Attach(p1, parapipe.NewPipeline(concurrency, func(msg int) (int, bool) {
		t.Error("Error expected to be skipped from processing, but worker has received one")
		return msg, true
	}))

	pipeline.Push(1)
	pipeline.Close()
	<-pipeline.Out()
}

func Benchmark1Pipe1Message(b *testing.B) {
	concurrency := 1
	pipeline := parapipe.NewPipeline(concurrency, func(msg int) (int, bool) { return msg, true })

	for n := 0; n < b.N; n++ {
		go func() {
			pipeline.Push(1)
		}()

		<-pipeline.Out()
	}

	pipeline.Close()
}

func Benchmark5Pipes1Message(b *testing.B) {
	concurrency := 1
	p1 := parapipe.NewPipeline(concurrency, func(msg int) (int, bool) { return msg, true })
	p2 := parapipe.Attach(p1, parapipe.NewPipeline(concurrency, func(msg int) (int, bool) { return msg, true }))
	p3 := parapipe.Attach(p2, parapipe.NewPipeline(concurrency, func(msg int) (int, bool) { return msg, true }))
	p4 := parapipe.Attach(p3, parapipe.NewPipeline(concurrency, func(msg int) (int, bool) { return msg, true }))
	pipeline := parapipe.Attach(p4, parapipe.NewPipeline(concurrency, func(msg int) (int, bool) { return msg, true }))

	for n := 0; n < b.N; n++ {
		go func() {
			pipeline.Push(1)
		}()

		<-pipeline.Out()
	}
}

func Benchmark1Pipe10000Messages(b *testing.B) {
	concurrency := 8
	pipeline := parapipe.NewPipeline(concurrency, func(msg int) (int, bool) { return msg, true })
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
	pipeline := parapipe.NewPipeline(concurrency, func(msg []int) ([]int, bool) { return msg, true })

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

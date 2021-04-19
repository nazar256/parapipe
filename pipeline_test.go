package parapipe

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestPipelineExecutesPipesInDefinedOrder(t *testing.T) {
	pipeline := NewPipeline(rand.Intn(5) + 2).
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
	concurrency := inputValuesCount
	pipeline := NewPipeline(concurrency).
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

	duration := time.Since(start)
	if duration > 150*time.Millisecond {
		t.Errorf(
			"Expected to be executed concurrently in 100ms, actual duration was %dms",
			duration/time.Millisecond,
		)
	}
}

func Benchmark1Pipe1Message(b *testing.B) {
	for n := 0; n < b.N; n++ {
		concurrency := 1
		pipeline := NewPipeline(concurrency).Pipe(func(msg interface{}) interface{} { return msg })
		genSmallestMessages(pipeline.In(), 1)
		for range pipeline.Out() {
		}
	}
}

func Benchmark5Pipes1Message(b *testing.B) {
	for n := 0; n < b.N; n++ {
		concurrency := 1
		pipeline := NewPipeline(concurrency).
			Pipe(func(msg interface{}) interface{} { return msg }).
			Pipe(func(msg interface{}) interface{} { return msg }).
			Pipe(func(msg interface{}) interface{} { return msg }).
			Pipe(func(msg interface{}) interface{} { return msg }).
			Pipe(func(msg interface{}) interface{} { return msg })
		genSmallestMessages(pipeline.In(), 1)
		for range pipeline.Out() {
		}
	}
}

func Benchmark1Pipe10000Messages(b *testing.B) {
	concurrency := 16
	for n := 0; n < b.N; n++ {
		pipeline := NewPipeline(concurrency).Pipe(func(msg interface{}) interface{} { return msg })
		genSmallestMessages(pipeline.In(), 10000)
		for range pipeline.Out() {
		}
	}
}
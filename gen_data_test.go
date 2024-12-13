package parapipe_test

import "github.com/nazar256/parapipe"

func makeRange(start, end int) []int {
	a := make([]int, end-start+1)
	for i := range a {
		a[i] = start + i
	}

	return a
}

func feedPipeline[T any](pipeline *parapipe.Pipeline[int, T], amount int) {
	go func() {
		for i := 0; i < amount; i++ {
			pipeline.Push(i)
		}

		pipeline.Close()
	}()
}

package parapipe_test

import "github.com/nazar256/parapipe"

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}

	return a
}

func feedPipeline(pipeline *parapipe.Pipeline, amount int) {
	go func() {
		for i := 0; i < amount; i++ {
			pipeline.Push(i)
		}
		pipeline.Close()
	}()
}

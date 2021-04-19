package parapipe

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func genIntMessages(dst chan<- interface{}, amount int) {
	go func() {
		for i := 0; i < amount; i++ {
			dst <- i
		}
		close(dst)
	}()
}

func genSmallestMessages(dst chan<- interface{}, amount int) {
	go func() {
		msg := struct{}{}
		for i := 0; i < amount; i++ {
			dst <- msg
		}
		close(dst)
	}()
}

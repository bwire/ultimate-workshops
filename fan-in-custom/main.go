package main

import (
	"fmt"
)

// fan-in without Waitgroup. Worker of workers concept.
func main() {
	numChannel := func(done <-chan interface{}, guard chan<- interface{}, name string) <-chan string {
		numbersChannel := make(chan string)

		go func() {
			defer close(numbersChannel)
			for _, n := range []int{1, 2, 3, 4, 5} {
				select {
				case <-done:
					return
				case numbersChannel <- fmt.Sprintf("%s-%d", name, n):
				}
			}
			guard <- struct{}{}
		}()

		return numbersChannel
	}

	funIn := func(done <-chan interface{}, guard <-chan interface{}, inputChans ...<-chan string) <-chan string {
		output := make(chan string)

		readInput := func(ic <-chan string) {
			for v := range ic {
				select {
				case <-done:
					return
				case output <- v:
				}
			}
		}

		// worker-of-workers concept
		go func() {
			for i := 0; i < len(inputChans); i++ {
				<-guard
			}
			close(output)
		}()

		for _, ch := range inputChans {
			go readInput(ch)
		}

		return output
	}

	done := make(chan interface{})
	workerGuard := make(chan interface{})

	defer close(done)
	defer close(workerGuard)

	merged := funIn(
		done,
		workerGuard,
		numChannel(done, workerGuard, "A"),
		numChannel(done, workerGuard, "B"),
		numChannel(done, workerGuard, "C"),
		numChannel(done, workerGuard, "D"),
		numChannel(done, workerGuard, "E"),
	)

	for v := range merged {
		fmt.Println(v)
	}
}

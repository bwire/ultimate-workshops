package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	numChannel := func(done <-chan interface{}, name string) <-chan string {
		numbersChannel := make(chan string)

		go func() {
			defer wg.Done()
			defer close(numbersChannel)
			for _, n := range []int{1, 2, 3, 4, 5} {
				select {
				case <-done:
					return
				case numbersChannel <- fmt.Sprintf("%s-%d", name, n):
				}
			}
		}()

		return numbersChannel
	}

	funIn := func(done <-chan interface{}, inputChans ...<-chan string) <-chan string {
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

		go func() {
			wg.Wait()
			close(output)
		}()

		for _, ch := range inputChans {
			go readInput(ch)
		}

		return output
	}

	done := make(chan interface{})
	defer close(done)

	wg.Add(3)
	merged := funIn(done, numChannel(done, "A"), numChannel(done, "B"), numChannel(done, "C"))

	for v := range merged {
		fmt.Println(v)
	}
}

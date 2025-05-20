package main

import "fmt"

func main() {
	inputFn := func(done <-chan interface{}, numbers ...int) <-chan int {
		numbersChan := make(chan int)

		go func() {
			defer close(numbersChan)

			for _, v := range numbers {
				select {
				case <-done:
					return
				case numbersChan <- v:
				}
			}
		}()

		return numbersChan
	}

	safeGoFn := func(f func()) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic processed")
			}
		}()

		f()
	}

	createWorker := func(done <-chan interface{}, guard chan<- interface{}, idx int, input <-chan int) <-chan int {
		worker := make(chan int)

		createFn := func() {
			defer close(worker)

			for {
				select {
				case <-done:
					return
				case v, ok := <-input:
					if !ok {
						select {
						case <-done:
							return
						case guard <- struct{}{}:
						}
					}
					select {
					case <-done:
					case worker <- v * v:
					}
				}
			}
		}

		go safeGoFn(createFn)
		return worker
	}

	fanOut := func(done <-chan interface{}, guard chan<- interface{}, input <-chan int, numChannels int) []<-chan int {
		workers := make([]<-chan int, 0, numChannels)

		for i := 0; i < numChannels; i++ {
			workers = append(workers, createWorker(done, guard, i+1, input))
		}

		return workers
	}

	fanIn := func(done <-chan interface{}, guard <-chan interface{}, workers ...<-chan int) <-chan int {
		output := make(chan int)

		readInput := func(ic <-chan int) {
			for v := range ic {
				select {
				case <-done:
					return
				case output <- v:
				}
			}
		}

		go func() {
			for i := 0; i < len(workers); i++ {
				<-guard
			}
			close(output)
		}()

		for _, w := range workers {
			go readInput(w)
		}

		return output
	}

	numChannels := 3
	done := make(chan interface{})
	guard := make(chan interface{})
	defer close(done)

	workers := fanOut(done, guard, inputFn(done, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10), numChannels)

	for v := range fanIn(done, guard, workers...) {
		fmt.Println(v)
	}

	close(guard)
}

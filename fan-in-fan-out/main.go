package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	// just a simple generator of integers from 0 to n...
	generate := func(done <-chan interface{}) chan int {
		seqIntChannel := make(chan int)

		go func() {
			defer close(seqIntChannel)

			for i:= 0;; i++ {
				select {
				case <-done:
					return
				case seqIntChannel <-i:
				}
			}
		}()

		return seqIntChannel
	}

	take := func(done <- chan interface{}, input <-chan int, n int) chan int {
		outputChannel := make(chan int)

		go func() {
			defer close(outputChannel)

			for i := 0; i < n; i++ {
				select {
				case <-done:
					return
				case v, ok := <-input:
					if !ok {
						return
					}
					select {
					case <-done:
						return
					case outputChannel <- v:
					}
				}
			}
		}()

		return outputChannel
	}

	// imitate some work
	someWork := func(val int, name string) int {
		fmt.Printf("processing value %v with channel %s\n", val, name)
		time.Sleep(time.Duration(rand.Intn(100) * int(time.Millisecond)))
		return  val + val
	}

	createWorker := func(
		done <-chan interface{},
		guard chan<- interface{},
		input <-chan int,
		name string,
	) <-chan int {
		worker := make(chan int)

		go func() {
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
							return
						}
					}
					select {
					case <-done:
						return
					case worker <- someWork(v, name):
					}
				}
			}
		}()

		return worker
	}

	fanOut := func(
		done <-chan interface{},
		guard chan<- interface{},
		input <-chan int,
		) []<-chan int {
		workers := make([]<-chan int, 0, 5)
		for _, name := range []string{"A", "B", "C", "D", "E" } {
			workers = append(workers, createWorker(done, guard, input, name))
		}
		return workers
	}

	fanIn := func(done <-chan interface{}, guard <-chan interface{}, inputChans ...<-chan int) <-chan int {
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
	defer close(done)

	workerGuard := make(chan interface{})

	inputs := fanOut(done, workerGuard, take(done, generate(done), 15))
	for v := range fanIn(done, workerGuard, inputs...) {
	 	fmt.Printf("output result: %v\n", v)
	}

	close(workerGuard)
}
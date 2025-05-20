package main

import (
	"context"
	"fmt"
)

func main() {
	inputFn := func(ctx context.Context, numbers ...int) <-chan int {
		numbersChan := make(chan int)

		go func() {
			defer close(numbersChan)

			for _, v := range numbers {
				select {
				case <-ctx.Done():
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

	createWorker := func(ctx context.Context, guard chan<- interface{}, input <-chan int) <-chan int {
		worker := make(chan int)

		createFn := func() {
			defer close(worker)

			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-input:
					if !ok {
						select {
						case <-ctx.Done():
							return
						case guard <- struct{}{}:
						}
					}
					select {
					case <-ctx.Done():
					case worker <- v * v:
					}
				}
			}
		}

		go safeGoFn(createFn)
		return worker
	}

	fanOut := func(ctx context.Context, guard chan<- interface{}, input <-chan int, numChannels int) []<-chan int {
		workers := make([]<-chan int, 0, numChannels)

		for i := 0; i < numChannels; i++ {
			workers = append(workers, createWorker(ctx, guard, input))
		}

		return workers
	}

	fanIn := func(ctx context.Context, guard <-chan interface{}, workers ...<-chan int) <-chan int {
		output := make(chan int)

		readInput := func(ic <-chan int) {
			for v := range ic {
				select {
				case <-ctx.Done():
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
	ctx, cancel := context.WithCancel(context.Background())

	guard := make(chan interface{})

	workers := fanOut(ctx, guard, inputFn(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10), numChannels)

	for v := range fanIn(ctx, guard, workers...) {
		fmt.Println(v)
	}

	cancel()
}

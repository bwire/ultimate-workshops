package main

import (
	"context"
	"fmt"
	"time"
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

	createWorker := func(ctx context.Context, guard chan<- interface{}, input <-chan int, idx int) <-chan int {
		worker := make(chan int)

		createFn := func() {
			defer func() {
				guard <- struct{}{}
				close(worker)
			}()

			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-input:
					if !ok {
						return
					}

					time.Sleep(time.Duration((idx+1) * 100) * time.Millisecond)

					select {
					case <-ctx.Done():
						return
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
			workers = append(workers, createWorker(ctx, guard, input, i))
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
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	fmt.Println("Start processing")

	guard := make(chan interface{})
	workers := fanOut(ctx, guard, inputFn(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10), numChannels)

	for v := range fanIn(ctx, guard, workers...) {
		fmt.Println(v)
	}

	fmt.Println("Done")
	close(guard)
}

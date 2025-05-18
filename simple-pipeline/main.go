package main

import (
	"fmt"
)

type StageFn[T any, R any] func(T) R

func stage[T any, R any](done <-chan interface{}, input <-chan T, stageFn StageFn[T, R]) <-chan R {
	output := make(chan R)

	go func() {
		defer close(output)

		for {
			select {
			case <-done:
			case v, ok := <-input:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case output <- stageFn(v):
				}
			}
		}
	}()

	return output
}

func main() {
	// just a simple generator of integers from 0 to n...
	generate := func(done <-chan interface{}) chan int {
		seqIntChannel := make(chan int)

		go func() {
			defer close(seqIntChannel)

			for i := 0; ; i++ {
				select {
				case <-done:
					return
				case seqIntChannel <- i:
				}
			}
		}()

		return seqIntChannel
	}

	take := func(done <-chan interface{}, input <-chan int, n int) chan int {
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

	done := make(chan interface{})
	defer close(done)

	input := take(done, generate(done), 10)
	multiplyBy2 := stage(done, input, func(val int) int { return val * 2 })
	plus100 := stage(done, multiplyBy2, func(val int) int { return val + 100 })
	toStringResult := stage(done, plus100, func(val int) string { return fmt.Sprintf("Result: %v", val) })

	for v := range toStringResult {
		fmt.Println(v)
	}
}

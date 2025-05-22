package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type errorWrapper struct {
	err  error
	name string
}

type flowStageFn func(int) (int, error)

type flowStage struct {
	n string
	f flowStageFn
}

func createStage(n string, f flowStageFn) flowStage {
	return flowStage{n: n, f: f}
}

func generate(ctx context.Context, numElements int) <-chan int {
	numStream := make(chan int)

	go func() {
		defer close(numStream)

		for i := 1; i <= numElements; i++ {
			select {
			case <-ctx.Done():
				return
			case numStream <- i:
			}
		}
	}()

	return numStream
}

func stage(
	ctx context.Context,
	input <-chan int,
	errorSignal chan<- errorWrapper,
	fs flowStage,
) <-chan int {
	output := make(chan int)

	go func() {
		defer func() {
			close(output)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-input:
				if !ok {
					return
				}

				v, err := fs.f(v)
				if err != nil {
					select {
					case <-ctx.Done():
						return
					case errorSignal <- errorWrapper{err: err, name: fs.n}:
						return
					}
				}

				select {
				case <-ctx.Done():
					return
				case output <- v:
					time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				}
			}
		}
	}()

	return output
}

func processFlow(
	ctx context.Context,
	errSignal chan<- errorWrapper,
	input <-chan int,
	stages ...flowStage,
) <-chan int {
	ic := input
	for _, s := range stages {
		ic = stage(ctx, ic, errSignal, s)
	}
	return ic
}

func main() {
	errSignal := make(chan errorWrapper)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// error tracker
	// here we guarantee, that errSignal will be closed either after reading the error
	// or after context is canceled.
	go func() {
		defer func() {
			close(errSignal)
			cancel()
		}()

		select {
		case <-ctx.Done():
			return
		case ew := <-errSignal:
			fmt.Printf("error occured in '%s': %v\n", ew.name, ew.err)
			return
		}
	}()

	num := 100

	withPossibleError := func(f flowStageFn, n int) flowStageFn {
		return func(v int) (int, error) {
			res, _ := f(v)
			if res > n {
				return 0, fmt.Errorf("this number %v is greater than %v", res, n)
			}
			return res, nil
		}
	}

	flow := processFlow(
		ctx,
		errSignal,
		generate(ctx, num),
		createStage("multiply by 2", func(v int) (int, error) { return v * 2, nil }),
		createStage("add 1", func(v int) (int, error) { return v + 1, nil }),
		createStage(
			"multiply value by value",
			withPossibleError(func(v int) (int, error) { return v * v, nil }, 1000),
		),
	)

	for v := range flow {
		fmt.Println(v)
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Println("flow finished")
}

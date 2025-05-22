package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type flowStageFn func(int) int

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

func stage(ctx context.Context, input <-chan int, fs flowStage) <-chan int {
	output := make(chan int)

	go func() {
		defer func() {
			fmt.Printf("stage '%s' stopped by timeout\n", fs.n)
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
				select {
				case <-ctx.Done():
					return
				case output <- fs.f(v):
					time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				}
			}
		}
	}()

	return output
}

func processFlow(ctx context.Context, input <-chan int, stages ...flowStage) <-chan int {
	ic := input
	for _, s := range stages {
		ic = stage(ctx, ic, s)
	}
	return ic
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	num := 100

	flow := processFlow(
		ctx,
		generate(ctx, num),
		createStage("multiply by 2", func(v int) int { return v * 2 }),
		createStage("add 1", func(v int) int { return v + 1 }),
		createStage("multiply value by value", func(v int) int { return v * v }),
	)

	for v := range flow {
		fmt.Println(v)
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Println("flow finished")
}

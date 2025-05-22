package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type transformNumber struct {
	n string
	f func(int) int
}

func generate(ctx context.Context, numElements int) <-chan int {
	numStream := make(chan int)

	go func() {
		defer func() {
			fmt.Println("generator stopped by timeout")
			close(numStream)
		}()

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

func stage(ctx context.Context, input <-chan int, tn transformNumber) <-chan int {
	output := make(chan int)

	go func() {
		defer func() {
			fmt.Printf("stage '%s' stopped by timeout\n", tn.n)
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
				case output <- tn.f(v):
					time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				}
			}
		}
	}()

	return output
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	num := 100

	inputSeq := generate(ctx, num)
	mulBy2 := stage(ctx, inputSeq, transformNumber{f: func(v int) int { return v * 2 }, n: "multiply by 2"})
	plus1 := stage(ctx, mulBy2, transformNumber{f: func(v int) int { return v + 1 }, n: "add 1"})
	valMulVal := stage(ctx, plus1, transformNumber{f: func(v int) int { return v * v }, n: "multiply value by value"})

	for v := range valMulVal {
		fmt.Println(v)
	}

	time.Sleep(200 * time.Millisecond)
}

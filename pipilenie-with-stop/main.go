package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func generate(ctx context.Context, numElements int) <-chan int {
	numStream := make(chan int)

	go func() {
		defer close(numStream)

		for i := 1; i <= numElements; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("input is closed by timeout")
				return
			case numStream <- i:
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			}
		}
	}()

	return numStream
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	num := 100

	for v := range generate(ctx, num) {
		fmt.Println(v)
	}
}

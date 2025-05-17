package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)

	worker := func(ctx context.Context, name string) {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("worker %s done\n", name)
				return
			case <-ticker.C:
				fmt.Printf("worker %s tick\n", name)
			}
		}
	}

	go worker(ctx, "A")
	go worker(ctx, "B")

	time.Sleep(2 * time.Second)
	cancel()
	fmt.Println("cancel called")

	time.Sleep(1 * time.Second)
	fmt.Println("exiting...")
}

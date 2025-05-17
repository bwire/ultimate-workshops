package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var wg sync.WaitGroup
	var once sync.Once

	worker := func(ctx context.Context, name string) {
		timer := time.NewTimer(time.Second * time.Duration(r.Intn(10)))
		defer timer.Stop()
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("worker %s done\n", name)
				return
			case <-timer.C:
				once.Do(func() {
					fmt.Printf("worker %s crashed\n", name)
					cancel()
				})
			}
		}
	}

	wg.Add(3)
	go worker(ctx, "A")
	go worker(ctx, "B")
	go worker(ctx, "C")

	wg.Wait()
	fmt.Println("all workers exited")
}

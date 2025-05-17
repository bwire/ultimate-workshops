package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	rootCtx := context.Background()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	g, ctx := errgroup.WithContext(rootCtx)

	worker := func(name string) error {
		timer := time.NewTimer(time.Second * time.Duration(r.Intn(10)))
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("worker %s done\n", name)
				return nil
			case <-timer.C:
				err := fmt.Errorf("worker %s crashed", name)
				return err
			}
		}
	}

	for _, name := range []string{"A", "B", "C"} {
		nm := name
		g.Go(func() error {
			return worker(nm)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Println("errgroup exited with error:", err)
	}

	fmt.Println("all workers exited")
}

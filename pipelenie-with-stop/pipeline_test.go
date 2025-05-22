package main

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestPipelineSuccessful(t *testing.T) {
	errChan := make(chan errorWrapper)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	flow := processFlow(
		ctx,
		errChan,
		generate(ctx, 5),
		createStage("x2", func(v int) (int, error) { return v * 2, nil }),
		createStage("+1", func(v int) (int, error) { return v + 1, nil }),
	)

	expectedCount := 5
	count := 0
	for range flow {
		count++
	}

	if count != expectedCount {
		t.Fatalf("expected %d results, got %d", expectedCount, count)
	}
}

func TestPipelineFailsOnError(t *testing.T) {
	errChan := make(chan errorWrapper)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	didFail := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
		case err := <-errChan:
			if err.err == nil {
				t.Errorf("expected error but got nil")
			}
			close(didFail)
		}
	}()

	flow := processFlow(
		ctx,
		errChan,
		generate(ctx, 100),
		createStage("too big", func(v int) (int, error) {
			if v > 10 {
				return 0, context.DeadlineExceeded
			}
			return v, nil
		}),
	)

	for range flow {
	}

	select {
	case <-didFail:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected error but timeout occurred")
	}
}

func TestPipelineTimeout(t *testing.T) {
	errChan := make(chan errorWrapper)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	flow := processFlow(
		ctx,
		errChan,
		generate(ctx, 1000),
		createStage("slow stage", func(v int) (int, error) {
			time.Sleep(10 * time.Millisecond)
			return v, nil
		}),
	)

	for range flow {
		// read to drain
	}

	// If test completes, it means timeout cancelled it — that's expected
}

func TestNoGoroutineLeak(t *testing.T) {
	before := runtime.NumGoroutine()

	errChan := make(chan errorWrapper)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	flow := processFlow(
		ctx,
		errChan,
		generate(ctx, 100),
		createStage("quick", func(v int) (int, error) { return v, nil }),
	)

	for range flow {
	}

	time.Sleep(200 * time.Millisecond)
	after := runtime.NumGoroutine()

	if diff := after - before; diff > 5 {
		t.Fatalf("potential goroutine leak detected: +%d goroutines", diff)
	}
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
)

func doSomethingUnreliable() error {
	if rand.Intn(10) < 7 {
		return errors.New("temporary failure")
	}
	return nil
}

func retryWithBackoffAndJitter(ctx context.Context, maxRetries int, baseDelay, maxDelay time.Duration) error {
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = doSomethingUnreliable()
		if err == nil {
			log.Printf("Attempt %d: success", attempt+1)
			return nil
		}

		if attempt == maxRetries-1 {
			break
		}

		backoff := baseDelay << attempt
		if backoff > maxDelay {
			backoff = maxDelay
		}
		jitter := time.Duration(rand.Int63n(int64(backoff) + 1))

		log.Printf("Attempt %d failed: err=%v, wait=%v (max backoff=%v)", attempt+1, err, jitter, backoff)
		timer := time.NewTimer(jitter)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	err := retryWithBackoffAndJitter(ctx, 5, 100*time.Millisecond, 2*time.Second)
	if err != nil {
		log.Printf("final result: %v", err)
		return
	}
	log.Println("final result: success")
}

// Fan-In: объединение потоков метрик с Alpha, Beta, Gamma.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Stage 1 — не изменять по условию задания.
func startServer(ctx context.Context, name string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(rand.Intn(500)) * time.Millisecond):
				out <- fmt.Sprintf("[%s] metric: %d", name, rand.Intn(100))
			}
		}
	}()
	return out
}

// FanIn сливает все входные каналы в один; закрывает выход после закрытия всех входов;
// учитывает отмену контекста.
func FanIn(ctx context.Context, inputs ...<-chan string) <-chan string {
	out := make(chan string)
	var wg sync.WaitGroup

	for _, in := range inputs {
		wg.Add(1)
		go func(ch <-chan string) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case val, ok := <-ch:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case out <- val:
					}
				}
			}
		}(in)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch1 := startServer(ctx, "Alpha")
	ch2 := startServer(ctx, "Beta")
	ch3 := startServer(ctx, "Gamma")
	ch4 := FanIn(ctx, ch1, ch2, ch3)

	for val := range ch4 {
		fmt.Println(val)
	}
}

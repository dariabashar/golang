// Унаследованный код с гонкой — для демонстрации `go run -race`.
// Ожидается: счётчик < 1000 и предупреждения race detector.
package main

import (
	"fmt"
	"sync"
)

func main() {
	var counter int
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++
		}()
	}
	wg.Wait()
	fmt.Println(counter)
}

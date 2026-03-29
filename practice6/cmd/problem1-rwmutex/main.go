// Thread-safe map: вариант 2 — обычный map + sync.RWMutex
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.RWMutex
	safeMap := make(map[string]int)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			mu.Lock()
			safeMap["key"] = key
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	mu.RLock()
	value := safeMap["key"]
	mu.RUnlock()
	fmt.Printf("Value: %d\n", value)
}

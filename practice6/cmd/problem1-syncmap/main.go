// Thread-safe map: вариант 1 — sync.Map (документация: https://pkg.go.dev/sync#Map)
package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var safeMap sync.Map

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			safeMap.Store("key", key)
		}(i)
	}
	wg.Wait()

	v, _ := safeMap.Load("key")
	fmt.Printf("Value: %v\n", v)
}

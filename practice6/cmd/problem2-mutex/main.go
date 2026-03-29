// Исправление гонки: sync.Mutex вокруг инкремента (каналы не используются).
//
// Почему без синхронизации счётчик не равен 1000: параллельные горутины
// выполняют неатомарное чтение–изменение–запись общей переменной counter,
// из‑за чего часть инкрементов теряется (lost update) и возникает data race.
package main

import (
	"fmt"
	"sync"
)

func main() {
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println(counter)
}

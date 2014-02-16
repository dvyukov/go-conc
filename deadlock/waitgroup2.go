// The program deadlocks.
// The opposite scheduling order is examined in waitgroup1.go.

package main

import (
	"sync"
	"time"
)

func main() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		time.Sleep(time.Second)
		mu.Lock()
		mu.Unlock()
		wg.Done()
	}()
	mu.Lock()
	wg.Wait()
	mu.Unlock()
}

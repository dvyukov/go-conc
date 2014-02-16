// Possible deadlock, if goroutines are scheduled in the opposite order,
// the program will deadlock.

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
		mu.Lock()
		mu.Unlock()
		wg.Done()
	}()
	time.Sleep(time.Second)
	mu.Lock()
	wg.Wait()
	mu.Unlock()
}

// NCASE=2

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
		if Case == 1 {
			time.Sleep(time.Second)
		}
		mu.Lock()
		wg.Done()
		mu.Unlock()
	}()
	if Case == 0 {
		time.Sleep(time.Second)
	}
	mu.Lock()
	wg.Wait()
	mu.Unlock()
}

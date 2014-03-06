package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	Case := Init(2)
	fmt.Printf("NODEADLOCK\n")

	var mu1 sync.Mutex
	var mu2 sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if Case == 1 {
			time.Sleep(time.Second)
		}
		mu1.Lock()
		mu2.Lock()
		mu2.Unlock()
		wg.Done()
		mu2.Lock()
		mu2.Unlock()
		mu1.Unlock()
	}()
	if Case == 0 {
		time.Sleep(time.Second)
	}
	mu1.Lock()
	mu2.Lock()
	mu2.Unlock()
	mu1.Unlock()
	wg.Wait()
	mu1.Lock()
	mu2.Lock()
	mu2.Unlock()
	mu1.Unlock()
}

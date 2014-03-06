package main

import (
	"fmt"
	"sync"
)

func main() {
	Init(1)
	fmt.Printf("DEADLOCK\n")

	c := make(chan int, 4)
	var mu sync.Mutex
	go func() {
		for i := 0; i < 10; i++ {
			mu.Lock()
			c <- i
			mu.Unlock()
		}
		close(c)
	}()
	for _ = range c {
		mu.Lock()
		mu.Unlock()
	}
}

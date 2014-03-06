package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	Case := Init(6)
	fmt.Printf("DEADLOCK\n")

	c := make(chan int, 1)
	c <- 1
	var mu sync.RWMutex
	done := make(chan bool)
	go func() {
		time.Sleep(sleep3[Case][0])
		mu.RLock()
		c <- 1
		mu.RUnlock()
		done <- true
	}()
	go func() {
		time.Sleep(sleep3[Case][1])
		mu.RLock()
		<-c
		mu.RUnlock()
		done <- true
	}()
	time.Sleep(sleep3[Case][2])
	mu.Lock()
	mu.Unlock()
	<-done
	<-done
}

package main

import (
	"fmt"
	"sync"
)

func main() {
	Init(1)
	fmt.Printf("NODEADLOCK\n")

	var mu1 sync.Mutex
	var mu2 sync.Mutex
	var once sync.Once
	done := make(chan bool)
	for p := 0; p < 2; p++ {
		go func() {
			mu1.Lock()
			mu2.Lock()
			mu2.Unlock()
			once.Do(func() {
				mu2.Lock()
				mu2.Unlock()
			})
			mu2.Lock()
			mu2.Unlock()
			mu1.Unlock()
			done <- true
		}()
	}
	for p := 0; p < 2; p++ {
		<-done
	}
}

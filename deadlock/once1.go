// NCASE=2

package main

import (
	"sync"
	"time"
)

func main() {
	var mu sync.Mutex
	var once sync.Once
	done := make(chan bool)
	go func() {
		if Case == 0 {
			time.Sleep(time.Second)
		}
		mu.Lock()
		once.Do(func() {
			mu.Lock()
			mu.Unlock()
		})
		mu.Unlock()
		done <- true
	}()
	if Case == 1 {
		time.Sleep(time.Second)
	}
	once.Do(func() {
		mu.Lock()
		mu.Unlock()
	})
	<-done
}

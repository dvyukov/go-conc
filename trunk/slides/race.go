package main

import (
	"fmt"
	"sync"
)

var mu0 sync.Mutex

// STARTMAIN OMIT
func main() {
	res := make(map[int]int)
	done := make(chan bool)

	// Spawn workers.
	for i := 0; i < 5; i++ {
		go func() {
			res[i] = i * i
			done <- true
		}()
	}

	// Join workers.
	for i := 0; i < 5; i++ {
		<-done
	}

	// Output results.
	fmt.Println(res)
}

// STOPMAIN OMIT

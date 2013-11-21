package main

import (
	"fmt"
	"time"
)

// STARTMAIN OMIT
func main() {
	i := 0
	timeout := time.After(time.Second)
loop:
	for {
		select {
		case <-timeout:
			break loop
		default:
			i = i + 1
		}
	}
	fmt.Println(i)
}

// STOPMAIN OMIT

package main

import (
	"fmt"
	"time"
)

func main() {
	Case := Init(2)
	fmt.Printf("DEADLOCK\n")

	i0 := make(chan int, 2)
	i1 := make(chan int, 2)
	o0 := make(chan int, 1)
	o1 := make(chan int, 1)
	done := make(chan bool)
	go func() {
		i0 <- 1
		i0 <- 1
		done <- true
	}()
	go func() {
		if Case == 1 {
			time.Sleep(time.Second)
		}
		i1 <- 1
		i1 <- 1
		done <- true
	}()
	go func() {
		for i := 0; i < 4; i++ {
			select {
			case <-o0:
			case v := <-i0:
				o1 <- v
			}
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 4; i++ {
			select {
			case <-o1:
			case v := <-i1:
				o0 <- v
			}
		}
		done <- true
	}()
	for p := 0; p < 4; p++ {
		<-done
	}
}

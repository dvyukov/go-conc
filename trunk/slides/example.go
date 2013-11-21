package main

import (
	"fmt"
)

// START OMIT
func main() {
	c := make(chan bool)
	m := make(map[string]string)
	go func() {
		m["1"] = "a"  // Racy
		c <- true
	}()
	for k, v := range m {
		fmt.Println(k, v) // Racy
	}
	<-c
}
// STOP OMIT

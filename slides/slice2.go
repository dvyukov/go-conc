package main

import "fmt"

// STARTMAIN OMIT
func main() {
	x := []int{0, 1, 2, 3, 4, 5}
	foo(x[1:3]) // HL
	fmt.Println(x)
}

func foo(x []int) {
	for i := range x {
		x[i] = -x[i]
	}
	_ = append(x, 42) // HL
}

// STOPMAIN OMIT

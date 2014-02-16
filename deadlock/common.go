package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

var Case int

func Init(ncase int) int {
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "usage: %v case\n", os.Args[0])
		os.Exit(1)
	}
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "%v\n", ncase)
		os.Exit(0)
	}
	c, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if c < 0 || c >= ncase {
		fmt.Fprintf(os.Stderr, "expect case to be within [0..%v) (got %v)\n", ncase, c)
		os.Exit(1)
	}
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Fprintf(os.Stderr, "failed to detect deadlock\n")
		os.Exit(0)
	}()
	return c
}

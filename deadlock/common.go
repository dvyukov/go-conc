package main

import (
	"os"
	"time"
	"strconv"
	"fmt"
)

var Case int

func init() {
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Fprintf(os.Stderr, "failed to detect deadlock\n")
		os.Exit(0)
	}()
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "usage: %v case\n", os.Args[0])
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		c, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if c < 0 || c > 100 {
			fmt.Fprintf(os.Stderr, "expect case to be within 0..100 (got %v)\n", c)
			os.Exit(1)
		}
		Case = c
	}
}

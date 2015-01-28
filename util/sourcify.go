package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

var (
	flagA = flag.Int("A", 0, "show that many context lines after match")
	flagB = flag.Int("B", 0, "show that many context lines before match")
	flagC = flag.Int("C", 0, "show that many context lines around match")

	re = regexp.MustCompile(`([\.,/,-,_,a-z,A-Z,0-9]+):([0-9]+)`)
)

func main() {
	flag.Parse()
	if *flagC != 0 {
		*flagA = *flagC
		*flagB = *flagC
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		processLine(out, s.Text())
	}
	if err := s.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse stdin: %v\n", err)
		os.Exit(1)
	}
}

func processLine(out *bufio.Writer, in string) {
	out.WriteString(in)
	out.WriteString("\n")
	matches := re.FindStringSubmatch(in)
	if len(matches) != 3 {
		return
	}
	ln, err := strconv.Atoi(matches[2])
	if err != nil || ln <= 0 {
		return
	}
	f, err := os.Open(matches[1])
	if err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for i := 1; s.Scan(); i++ {
		if i >= ln-*flagB && i <= ln+*flagA {
			out.WriteString(s.Text())
			out.WriteString("\n")
			if i == ln+*flagA {
				break
			}
		}
	}
}

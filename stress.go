package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"syscall"
)

var (
	flagN       = flag.Int("n", 1, "number of parallel processes")
	flagTimeout = flag.Duration("timeout", time.Hour, "timeout for each process")
	flagKill    = flag.Bool("kill", true, "kill timed out processes, or just print pid")
	flagLogPath = flag.String("logpath", "", "path prefix for log files")
)

func main() {
	flag.Parse()
	if *flagN <= 0 || *flagTimeout <= 0 || len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	res := make(chan []byte)
	for i := 0; i < *flagN; i++ {
		go func() {
			for {
				cmd := exec.Command(flag.Args()[0], flag.Args()[1:]...)
				done := make(chan bool)
				if *flagTimeout > 0 {
					go func() {
						select {
						case <-done:
							return
						case <-time.After(*flagTimeout):
						}
						if !*flagKill {
							fmt.Printf("process timed out %d\n", cmd.Process.Pid)
							return
						}
						cmd.Process.Signal(syscall.SIGABRT)
						select {
						case <-done:
							return
						case <-time.After(10*time.Second):
						}
						cmd.Process.Kill()
					}()
				}
				out, err := cmd.CombinedOutput()
				close(done)
				if err != nil {
					out = append(out, fmt.Sprintf("\n\nERROR: %v\n", err)...)
				} else {
					out = []byte{}
				}
				res <- out
			}
		}()
	}
	if *flagLogPath == "" {
		*flagLogPath = filepath.Join(os.TempDir(), "stress.")
	}
	n := 0
	ticker := time.NewTicker(5 * time.Second).C
	for {
		select {
		case out := <-res:
			n++
			if len(out) == 0 {
				continue
			}
			f, err := os.Create(fmt.Sprintf("%s%v", *flagLogPath, time.Now().UnixNano()))
			if err != nil {
				fmt.Printf("\n%s\n%s\n", err, out)
			} else {
				f.Write(out)
				f.Close()
				if len(out) > 2<<10 {
					out = out[:2<<10]
				}
				fmt.Printf("\n%s\n%s\n", f.Name(), out)
			}
		case <-ticker:
			fmt.Printf("%v runs so far\n", n)
		}
	}
}

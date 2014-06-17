package main

import (
	"net/http"
	"time"
)

func main() {
	_ = multiGet([]string{"aa", "bb"})
	_ = multiGetTimeout([]string{"aa", "bb"}, time.Second)
}

func multiGet(urls []string) (results []*http.Response) {
	res := make(chan *http.Response)
	for _, url := range urls {
		go func(url string) {
			resp, _ := http.Get(url)
			res <- resp
		}(url)
	}
	for _ = range urls {
		results = append(results, <-res)
	}
	return
}

func multiGetTimeout(urls []string, timeout time.Duration) (results []*http.Response) {
	res := make(chan *http.Response)
	for _, url := range urls {
		go func(url string) {
			resp, _ := http.Get(url)
			res <- resp
		}(url)
	}
	t := time.After(timeout)
	for _ = range urls {
		select {
		case resp := <-res:
			results = append(results, resp)
		case <-t:
			return
		}
	}
	return
}

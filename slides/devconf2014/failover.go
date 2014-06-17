package main

import (
	"net/http"
	"time"
)

func main() {
	_ = getFailover("aa", "bb")
}

func requestServer(url string) *http.Response {
	return nil
}

func getFailover(primary, secondary string) *http.Response {
	res := make(chan *http.Response, 1)
	go func() {
		res <- requestServer(primary)
	}()
	select {
	case resp := <-res:
		return resp
	case <-time.After(50 * time.Millisecond):
	}
	go func() {
		res <- requestServer(secondary)
	}()
	return <-res
}

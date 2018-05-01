package main

import (
	"log"
	"net/http"
	"os"
	"sync"
)

// mischief
type loki struct {
	sync.Mutex
	callcount int
}

// this is just a function that will return true
// the last 5 out of every 15 times called
func (l *loki) shouldBreakThings() bool {
	l.Lock()
	defer l.Unlock()
	l.callcount++
	if l.callcount > 15 {
		l.callcount = 0
		return false
	}
	if l.callcount > 10 {
		return true
	}

	return false
}

var (
	l      *loki
	logger *log.Logger
)

func init() {
	l = &loki{}
	logger = log.New(os.Stderr, "dependency: ", 0)
}

func handleRequest(rw http.ResponseWriter, r *http.Request) {
	// ignore favicon
	if r.URL.Path == "/favicon.ico" {
		rw.WriteHeader(http.StatusOK)
		return
	}
	if l.shouldBreakThings() {
		logger.Print("👎")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.Print("👍")
	rw.WriteHeader(http.StatusOK)
	return
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe("0.0.0.0:8081", nil)
}

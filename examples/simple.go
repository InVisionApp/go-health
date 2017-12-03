package main

import (
	"net/http"
)

func handler(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("go-health is awesome!"))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

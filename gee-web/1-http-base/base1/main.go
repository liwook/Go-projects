package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	//可以写成匿名函数(lambda表达式),handler echoes r.URL.Path
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	})

	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe("localhost:10000", nil))
}

// handler echoes r.URL.Header
func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
}

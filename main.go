package main

import (
	"fmt"
	"net/http"
	"shawg/shawg"
)

func main() {
	r := shawg.New()
	r.GET("/hello", handleHello)
	r.Run(":9091")
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello")
}

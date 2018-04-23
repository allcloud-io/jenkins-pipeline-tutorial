package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	version = "1.0"
)

func main() {
	h, _ := os.Hostname()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<h1>It works!</h1>\n<h1>Hostname: %s</h1>\n<h1>Version: %s</h1>", h, version)
	})

	log.Fatal(http.ListenAndServe(":80", nil))
}

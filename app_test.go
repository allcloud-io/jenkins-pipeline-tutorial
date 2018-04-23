package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	go func() {
		main()
	}()

	res, err := http.Get("http://localhost")
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	h, _ := os.Hostname()
	want := fmt.Sprintf("<h1>It works!</h1>\n<h1>Hostname: %s</h1>\n<h1>Version: %s</h1>", h, version)
	got := string(data)

	if got != want {
		t.Fatalf("Got wrong response: got %v want %v", got, want)
	}
}

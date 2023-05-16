package main

import (
	"net/http"
	"testing"
)

func TestMain(t *testing.T) {
	go main()

	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		t.Errorf("failed to create GET request: %v", err)
	}

	want := http.StatusNotFound
	if resp != nil && resp.StatusCode != want {
		t.Errorf("want: %d - got: %d", want, resp.StatusCode)
	}
}

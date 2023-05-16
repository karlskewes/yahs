package main

import (
	"log"
	"net/http"
)

func main() {
	log.Print("go-yahs starting")

	mux := http.NewServeMux()
	mux.Handle("/", http.NotFoundHandler())

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Print(err)
	}
}

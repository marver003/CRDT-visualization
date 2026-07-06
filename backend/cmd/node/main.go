package main

import (
	"log"
	"net/http"

	Store "github.com/marver003/crdt/internal/store"
	HttpTransport "github.com/marver003/crdt/internal/transport/http"
)

func main() {
	store := Store.New()
	handler := HttpTransport.NewHandler(store)
	router := HttpTransport.NewRouter(handler)

	addr := ":8080"
	log.Printf("GCounter API listening on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
package main

import (
	"log"
	"net/http"

	"github.com/marver003/crdt/internal/store"
	nodehttp "github.com/marver003/crdt/internal/transport/http/node"
)

func main() {
	store := store.New()
	handler := nodehttp.NewHandler(store)
	router := nodehttp.NewRouter(handler)

	addr := ":8080"
	log.Printf("GCounter API listening on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

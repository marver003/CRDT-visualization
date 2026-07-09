package main

import (
	"log"
	"net/http"

	"github.com/marver003/crdt/internal/simulator"
	simulatorhttp "github.com/marver003/crdt/internal/transport/http/simulator"
)

func main() {

	registry := simulator.NewRegistry()
	handler := simulatorhttp.NewHandler(registry)
	router := simulatorhttp.NewRouter(handler)

	addr := ":8080"
	log.Printf("GCounter API listening on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

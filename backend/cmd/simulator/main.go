package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	simhttp "github.com/marver003/crdt/internal/transport/http/simulator"
)

func main() {

	ptrHttpPort := flag.Int("phttp", 8080, "http simulator server port")

	flag.Parse()

	httpAddr := fmt.Sprintf(":%d", *ptrHttpPort)

	store := simhttp.NewStore()
	handler := simhttp.NewHandler(store)
	router := simhttp.NewRouter(handler)

	log.Printf("Simulator API listening on %d", *ptrHttpPort)

	//go grpc.GossipServer(grpcAddr, node.Counter)
	//go node.GossipService.Gossip()

	if err := http.ListenAndServe(httpAddr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

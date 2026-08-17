package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	simhttp "github.com/marver003/crdt/internal/transport/http/simulator"
	"github.com/marver003/crdt/internal/simulator"
)

func main() {

	ptrHttpPort := flag.Int("phttp", 8080, "http simulator server port")

	flag.Parse()

	httpAddr := fmt.Sprintf(":%d", *ptrHttpPort)

	sim := simulator.New()
	handler := simhttp.NewHandler(sim)
	router := simhttp.NewRouter(handler)

	log.Printf("Simulator API listening on %d", *ptrHttpPort)

	//go grpc.GossipServer(grpcAddr, node.Counter)
	//go node.GossipService.Gossip()

	if err := http.ListenAndServe(httpAddr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

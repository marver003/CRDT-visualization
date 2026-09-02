package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/marver003/crdt/internal/simulator"
	simhttp "github.com/marver003/crdt/internal/transport/http/simulator"
)

func ParseFlags() (int, error) {
	var httpPort int

	flag.IntVar(&httpPort, "phttp", 8080, "The port used for HTTP server of simulator")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if httpPort < 1 || httpPort > 65535 {
		return 0, fmt.Errorf("-phttp must be between 1 and 65535")
	}

	return httpPort, nil
}

func main() {

	httpPort, err := ParseFlags()

	if err != nil {
		log.Fatalln(err)
	}

	httpAddr := fmt.Sprintf(":%d", httpPort)

	sim := simulator.New()
	handler := simhttp.NewHandler(sim)
	router := simhttp.NewRouter(handler)

	log.Printf("Simulator API listening on %d", httpPort)

	//go grpc.GossipServer(grpcAddr, node.Counter)
	//go node.GossipService.Gossip()

	if err := http.ListenAndServe(httpAddr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

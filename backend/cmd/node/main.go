package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/marver003/crdt/internal/node"
	"github.com/marver003/crdt/internal/transport/grpc"
	nodehttp "github.com/marver003/crdt/internal/transport/http/live"
	"github.com/marver003/crdt/internal/types"
)

func parseNeighborString(neighborString string) []types.Peer {

	neighborStringArray := strings.Split(neighborString, ";")

	var neighbors []types.Peer

	for i, neighborString := range neighborStringArray {
		peer := types.Peer{
			ID:       types.ReplicaID(fmt.Sprint('A' + i)),
			GRPCAddr: neighborString,
		}
		neighbors = append(neighbors, peer)
	}

	return neighbors
}

type FlagConfig struct {
	NodeID    string
	HTTPPort  int
	GRPCPort  int
	Neighbors string
}

func ParseFlags() (FlagConfig, error) {
	var (
		nodeID    string
		httpPort  int
		grpcPort  int
		neighbors string
	)

	flag.StringVar(&nodeID, "id", "", "Unique positive number that represents node ID.")
	flag.StringVar(&neighbors, "n", "", "Semicoln-separated list of other node gRPC server addresses (e.g.: localhost:9081;localhost:9082).")
	flag.IntVar(&grpcPort, "pgrpc", 9080, "The port used for gRPC server of this node.")
	flag.IntVar(&httpPort, "phttp", 8080, "The port used for HTTP server of this node.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if nodeID == "A" {
		return FlagConfig{}, fmt.Errorf("-id is required")
	}

	if httpPort < 1 || httpPort > 65535 {
		return FlagConfig{}, fmt.Errorf("-phttp must be between 1 and 65535")
	}

	if grpcPort < 1 || grpcPort > 65535 {
		return FlagConfig{}, fmt.Errorf("-pgrpc must be between 1 and 65535")
	}

	return FlagConfig{
		NodeID:    nodeID,
		HTTPPort:  httpPort,
		GRPCPort:  grpcPort,
		Neighbors: neighbors,
	}, nil
}

func main() {

	flags, err := ParseFlags()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err)
		flag.Usage()
		os.Exit(2)
	}

	httpAddr := fmt.Sprintf(":%d", flags.HTTPPort)
	grpcAddr := fmt.Sprintf(":%d", flags.GRPCPort)
	peers := parseNeighborString(flags.Neighbors)

	node := node.Create(types.ReplicaID(flags.NodeID), peers)
	handler := nodehttp.NewHandler(node)
	router := nodehttp.NewRouter(handler)

	log.Printf("GCounter API listening on %d", flags.HTTPPort)

	go grpc.GossipServer(grpcAddr, node.Counter)
	go node.GossipService.Gossip()

	if err := http.ListenAndServe(httpAddr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

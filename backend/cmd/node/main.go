package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/marver003/crdt/internal/node"
	"github.com/marver003/crdt/internal/transport/grpc"
	nodehttp "github.com/marver003/crdt/internal/transport/http/node"
	"github.com/marver003/crdt/internal/types"
)

func parseNeighborString(neighborString string) []types.Peer {

	neighborStringArray := strings.Split(neighborString, ";")

	var neighbors []types.Peer

	for i, neighborString := range neighborStringArray {
		peer := types.Peer{
			ID:       types.ReplicaID(i),
			GRPCAddr: neighborString,
		}
		neighbors = append(neighbors, peer)
	}

	return neighbors
}

func main() {

	ptrNodeId := flag.Uint64("id", 1, "node id (MUST BE UNIQUE)")
	ptrHttpPort := flag.Int("phttp", 8080, "http server port")
	ptrGrpcPort := flag.Int("pgrpc", 9080, "grpc server port")
	ptrNeighbors := flag.String("n", "", "neighbors")

	flag.Parse()

	httpAddr := fmt.Sprintf(":%d", *ptrHttpPort)
	grpcAddr := fmt.Sprintf(":%d", *ptrGrpcPort)
	peers := parseNeighborString(*ptrNeighbors)

	node := node.Create(types.ReplicaID(*ptrNodeId), peers)
	handler := nodehttp.NewHandler(node)
	router := nodehttp.NewRouter(handler)

	log.Printf("GCounter API listening on %d", *ptrHttpPort)

	go grpc.GossipServer(grpcAddr, node.Counter)
	go node.GossipService.Gossip()

	if err := http.ListenAndServe(httpAddr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

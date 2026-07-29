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

	peer1 := types.Peer{
		ID:       1,
		GRPCAddr: neighborStringArray[0],
	}

	peer2 := types.Peer{
		ID:       2,
		GRPCAddr: neighborStringArray[1],
	}

	return []types.Peer{peer1, peer2}
}

func main() {

	ptrHttpPort := flag.Int("phttp", 8080, "http server port")
	ptrGrpcPort := flag.Int("pgrpc", 9080, "grpc server port")
	ptrNeighbors := flag.String("n", "", "neighbors")

	flag.Parse()

	httpAddr := fmt.Sprintf(":%d", *ptrHttpPort)
	grpcAddr := fmt.Sprintf(":%d", *ptrGrpcPort)
	peers := parseNeighborString(*ptrNeighbors)

	node := node.Create(0, peers)
	handler := nodehttp.NewHandler(node)
	router := nodehttp.NewRouter(handler)

	log.Printf("GCounter API listening on %d", *ptrHttpPort)

	go grpc.GossipServer(grpcAddr, node.Counter)
	go node.GossipService.Gossip()

	if err := http.ListenAndServe(httpAddr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

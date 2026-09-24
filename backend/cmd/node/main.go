package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/marver003/crdt/internal/config"
	"github.com/marver003/crdt/internal/node"
	"github.com/marver003/crdt/internal/transport/grpc"
	nodehttp "github.com/marver003/crdt/internal/transport/http/live"
	"github.com/marver003/crdt/internal/types"
)

func main() {
	configPath := flag.String("config", "", "Path to the live YAML config shared by all nodes.")
	nodeId := flag.String("id", "", "Id of this node, as defined under nodes in the config.")
	flag.Parse()

	if *configPath == "" || *nodeId == "" {
		fmt.Fprintln(os.Stderr, "error: --config and --id are required")
		flag.Usage()
		os.Exit(2)
	}

	cfg, err := config.LoadLive(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	id := types.ReplicaId(*nodeId)

	nodeCfg, exists := cfg.Node(id)
	if !exists {
		log.Fatalf("config: node %q is not defined in %s", id, *configPath)
	}

	n, err := node.Create(id, nodeCfg.Crdt, cfg.Peers(id))
	if err != nil {
		log.Fatalf("node creation failed: %v", err)
	}

	handler := nodehttp.NewHandler(n)
	router := nodehttp.NewRouter(handler)

	log.Printf("Node %s (%s) API listening on %d, gRPC on %d", id, nodeCfg.Crdt, nodeCfg.HTTPPort, nodeCfg.GRPCPort)

	go grpc.GossipServer(fmt.Sprintf(":%d", nodeCfg.GRPCPort), n.Crdt)
	go n.GossipService.Gossip()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", nodeCfg.HTTPPort), router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

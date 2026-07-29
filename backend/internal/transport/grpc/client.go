package grpc

import (
	"context"
	"log"
	"time"

	"github.com/marver003/crdt/api/grpc/replication"
	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GossipService struct {
	peers    []types.Peer
	counter  *crdt.GCounter
	interval time.Duration
}

func GossipClient(peers []types.Peer, counter *crdt.GCounter, interval time.Duration) *GossipService {
	log.Print("Create gossip service")
	return &GossipService{
		peers:    peers,
		counter:  counter,
		interval: interval,
	}
}

func (g *GossipService) Gossip() {
	log.Print("Start gossiping...")
	var connections []*grpc.ClientConn
	var grpcClients []replication.ReplicationServiceClient
	for _, peer := range g.peers {

		conn, err := grpc.NewClient(peer.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			panic(err)
		}

		connections = append(connections, conn)
		grpcClients = append(grpcClients, replication.NewReplicationServiceClient(conn))

	}

	contextRepl, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		time.Sleep(g.interval)
		for index, client := range grpcClients {
			log.Printf("Gossip to %s\n", g.peers[index].GRPCAddr)
			stateBytes, err := g.counter.Marshal()
			if err != nil {
				panic(err)
			}
			request := replication.PushStateRequest{
				CrdtId:   0,
				CrdtType: "GCounter",
				State:    stateBytes,
			}
			if _, err := client.PushState(contextRepl, &request); err != nil {
				log.Fatalf("PushState failed: %v", err.Error())
			}
		}
	}
}

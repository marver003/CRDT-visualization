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

const timeout = 10 * time.Second

type GossipService struct {
	id       types.ReplicaId
	peers    []types.Peer
	crdt     crdt.CRDT
	interval time.Duration
	changed  chan struct{}
}

func GossipClient(id types.ReplicaId, peers []types.Peer, crdtInstance crdt.CRDT, interval time.Duration) *GossipService {
	return &GossipService{
		id:       types.ReplicaId(id),
		peers:    peers,
		crdt:     crdtInstance,
		interval: interval,
		changed:  make(chan struct{}, 1),
	}
}

func (g *GossipService) Notify() {
	select {
	case g.changed <- struct{}{}:
	default:
	}
}

func (g *GossipService) Gossip() {
	log.Println("Gossip service started.")

	var triggers []chan struct{}

	for _, peer := range g.peers {
		conn, err := grpc.NewClient(peer.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Connection to gRPC server %s failed: %v\n", peer.GRPCAddr, err)
			continue
		}

		trigger := make(chan struct{}, 1)
		triggers = append(triggers, trigger)

		go g.pushLoop(peer.GRPCAddr, replication.NewReplicationServiceClient(conn), trigger)
	}

	ticker := time.NewTicker(g.interval)

	for {
		select {
		case <-g.changed:
		case <-ticker.C:
		}

		for _, trigger := range triggers {
			select {
			case trigger <- struct{}{}:
			default:
			}
		}
	}
}

func (g *GossipService) pushLoop(addr string, client replication.ReplicationServiceClient, trigger <-chan struct{}) {
	for {
		<-trigger

		log.Printf("Gossip to %s\n", addr)

		stateBytes, err := g.crdt.Marshal()
		if err != nil {
			log.Fatalf("Marshal failed: %v", err.Error())
		}

		request := replication.PushStateRequest{
			CrdtId:   string(g.id),
			CrdtType: string(g.crdt.Type()),
			State:    stateBytes,
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		_, err = client.PushState(ctx, &request)
		cancel()

		if err != nil {
			log.Printf("PushState to %s failed: %v\n", addr, err.Error())
		}
	}
}

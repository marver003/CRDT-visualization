package node

import (
	"log"
	"time"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/transport/grpc"
	"github.com/marver003/crdt/internal/types"
)

type Node struct {
	ID            types.ReplicaID
	Counter       *crdt.GCounter
	Peers         []types.Peer
	GossipService *grpc.GossipService
}

// Create initializes and returns a node
func Create(id types.ReplicaID, peers []types.Peer) *Node {
	log.Println("Creating new node")
	counter := crdt.NewGCounter(id)
	return &Node{
		ID:            id,
		Counter:       counter,
		Peers:         peers,
		GossipService: grpc.GossipClient(peers, counter, 5*time.Second),
	}
}

// ListSnapshot returns a snapshot of replica's state
func (n *Node) ListSnapshot() crdt.GCounterSnapshot {
	return n.Counter.GetSnapshot()
}

// GetValue returns sum of counter counts
func (n *Node) GetValue() uint64 {
	return n.Counter.Value()
}

func (n *Node) Increment() {
	n.Counter.Increment()
}

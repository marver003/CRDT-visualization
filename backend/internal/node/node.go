package node

import (
	"log"
	"time"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/transport/grpc"
	"github.com/marver003/crdt/internal/types"
)

type Node struct {
	Id            types.ReplicaId
	Counter       *crdt.GCounter
	Peers         []types.Peer
	GossipService *grpc.GossipService
}

// Create initializes and returns a node
func Create(id types.ReplicaId, peers []types.Peer) *Node {
	log.Println("Creating new node")
	counter := crdt.NewGCounter(id)
	return &Node{
		Id:            id,
		Counter:       counter,
		Peers:         peers,
		GossipService: grpc.GossipClient(id, peers, counter, 5*time.Second),
	}
}

func CreateSimNode(id types.ReplicaId) *Node {
	log.Println("Creating new SIM node")
	counter := crdt.NewGCounter(id)
	return &Node{
		Id:      id,
		Counter: counter,
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

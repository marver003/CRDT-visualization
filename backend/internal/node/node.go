package node

import (
	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/types"
)

type Node struct {
	ID      types.ReplicaID
	Counter *crdt.GCounter
	Peers   []types.Peer
}

// Create initializes and returns a node
func Create(id types.ReplicaID, peers []types.Peer) *Node {
	return &Node{
		ID:      id,
		Counter: crdt.NewGCounter(id),
		Peers:   peers,
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

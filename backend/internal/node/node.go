package node

import (
	"fmt"
	"log"
	"time"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/transport/grpc"
	"github.com/marver003/crdt/internal/types"
)

type Node struct {
	Id            types.ReplicaId
	Crdt          crdt.CRDT
	Peers         []types.Peer
	GossipService *grpc.GossipService
}

// Create initializes and returns a node
func Create(id types.ReplicaId, crdtType types.CrdtType, peers []types.Peer) (*Node, error) {
	log.Println("Creating new node")
	crdtInstance, err := crdt.NewCrdt(id, crdtType)
	if err != nil {
		return nil, err
	}

	return &Node{
		Id:            id,
		Crdt:          crdtInstance,
		Peers:         peers,
		GossipService: grpc.GossipClient(id, peers, crdtInstance, 5*time.Second),
	}, nil
}

func CreateSimNode(id types.ReplicaId, crdtType types.CrdtType) (*Node, error) {
	log.Println("Creating new SIM node")
	crdtInstance, err := crdt.NewCrdt(id, crdtType)
	if err != nil {
		return nil, err
	}

	return &Node{
		Id:   id,
		Crdt: crdtInstance,
	}, nil
}

// ListSnapshot returns a snapshot of replica's state
func (n *Node) ListSnapshot() crdt.Snapshot {
	return n.Crdt.GetSnapshot()
}

// Increment is GCounter specific
func (n *Node) Increment() error {
	gcounter, ok := n.Crdt.(*crdt.GCounter)

	if !ok {
		return fmt.Errorf("Unsuccessful assertion to *crdt.GCounter")
	}

	gcounter.Increment()

	return nil
}

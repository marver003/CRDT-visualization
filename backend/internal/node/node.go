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

func (n *Node) Add(element string) error {
	gset, ok := n.Crdt.(*crdt.GSet)

	if !ok {
		return fmt.Errorf("Unsuccessful assertion to *crdt.GSet")
	}

	gset.Add(element)

	return nil
}

func (n *Node) IncrementBy(amount uint64) error {
	gcounter, ok := n.Crdt.(*crdt.GCounter)

	if !ok {
		return fmt.Errorf("Unsuccessful assertion to *crdt.GCounter")
	}

	gcounter.Add(amount)

	return nil
}

func (n *Node) GenerateOp(amount uint64) (crdt.CounterOp, error) {
	counter, ok := n.Crdt.(*crdt.GCounterOp)

	if !ok {
		return crdt.CounterOp{}, fmt.Errorf("Unsuccessful assertion to *crdt.GCounterOp")
	}

	return counter.Generate(amount), nil
}

func (n *Node) ApplyOp(op crdt.CounterOp) (bool, error) {
	counter, ok := n.Crdt.(*crdt.GCounterOp)

	if !ok {
		return false, fmt.Errorf("Unsuccessful assertion to *crdt.GCounterOp")
	}

	return counter.Apply(op), nil
}

func (n *Node) Have() (map[types.ReplicaId]uint64, error) {
	counter, ok := n.Crdt.(*crdt.GCounterOp)

	if !ok {
		return nil, fmt.Errorf("Unsuccessful assertion to *crdt.GCounterOp")
	}

	return counter.Have(), nil
}

func (n *Node) OpsAfter(seq uint64) ([]crdt.CounterOp, error) {
	counter, ok := n.Crdt.(*crdt.GCounterOp)

	if !ok {
		return nil, fmt.Errorf("Unsuccessful assertion to *crdt.GCounterOp")
	}

	return counter.OpsAfter(seq), nil
}

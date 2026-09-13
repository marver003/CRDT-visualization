package simulator

import (
	"fmt"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/node"
	"github.com/marver003/crdt/internal/types"
)

type Store struct {
	Nodes map[types.ReplicaId]*node.Node
}

func NewStore() *Store {
	return &Store{make(map[types.ReplicaId]*node.Node)}
}

func (s *Store) CreateNode(id types.ReplicaId) error {

	if _, exists := s.Nodes[id]; exists {
		return fmt.Errorf("node %q already exists", id)
	}

	s.Nodes[id] = node.CreateSimNode(id)

	return nil
}

func (s *Store) CreateNodeWithSnapshot(id types.ReplicaId, snapshot crdt.GCounterSnapshot) {
	s.Nodes[id] = node.CreateSimNode(id)
	s.Nodes[id].Counter.ApplySnapshot(snapshot)
}

func (s *Store) RemoveNode(id types.ReplicaId) error {
	if _, exists := s.Nodes[id]; !exists {
		return fmt.Errorf("node %q does not exist", id)
	}

	delete(s.Nodes, id)

	return nil
}

func (s *Store) GetState() map[types.ReplicaId]map[types.ReplicaId]uint64 {
	state := make(map[types.ReplicaId]map[types.ReplicaId]uint64, len(s.Nodes))

	for id, node := range s.Nodes {
		state[id] = node.Counter.GetSnapshotCounts().Counts
	}

	return state
}

func (s *Store) IncrementNodeCounter(id types.ReplicaId) (crdt.GCounterSnapshot, error) {
	if _, exists := s.Nodes[id]; !exists {
		return crdt.GCounterSnapshot{}, fmt.Errorf("node %q does not exist", id)
	}

	s.Nodes[id].Counter.Increment()

	return s.Nodes[id].Counter.GetSnapshot(), nil
}

func (s *Store) MergeNodeStates(sourceId, targetId types.ReplicaId) (crdt.GCounterSnapshot, error) {
	if _, exists := s.Nodes[sourceId]; !exists {
		return crdt.GCounterSnapshot{}, fmt.Errorf("source node %q does not exist", sourceId)
	}

	if _, exists := s.Nodes[targetId]; !exists {
		return crdt.GCounterSnapshot{}, fmt.Errorf("target node %q does not exist", targetId)
	}

	SourceCounterSnapshot := s.Nodes[sourceId].Counter.GetSnapshot()

	s.Nodes[targetId].Counter.Merge(&SourceCounterSnapshot)

	return s.Nodes[targetId].Counter.GetSnapshot(), nil
}

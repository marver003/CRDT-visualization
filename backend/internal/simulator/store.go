package simulator

import (
	"fmt"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/node"
	"github.com/marver003/crdt/internal/types"
)

type Store struct {
	nodes map[types.ReplicaID]*node.Node
}

func NewStore() *Store {
	return &Store{make(map[types.ReplicaID]*node.Node)}
}

func (s *Store) CreateNode(id types.ReplicaID) error {

	if _, exists := s.nodes[id]; exists {
		return fmt.Errorf("node %d already exists", id)
	}

	s.nodes[id] = node.CreateSimNode(id)

	return nil
}

func (s *Store) RemoveNode(id types.ReplicaID) error {
	if _, exists := s.nodes[id]; !exists {
		return fmt.Errorf("node %d does not exist", id)
	}

	delete(s.nodes, id)

	return nil
}

func (s *Store) GetState() map[types.ReplicaID]map[types.ReplicaID]uint64 {
	state := make(map[types.ReplicaID]map[types.ReplicaID]uint64, len(s.nodes))

	for id, node := range s.nodes {
		state[id] = node.Counter.GetSnapshotCounts().Counts
	}

	return state
}

func (s *Store) IncrementNodeCounter(id types.ReplicaID) (crdt.GCounterSnapshot, error) {
	if _, exists := s.nodes[id]; !exists {
		return crdt.GCounterSnapshot{}, fmt.Errorf("node %d does not exist", id)
	}

	s.nodes[id].Counter.Increment()

	return s.nodes[id].Counter.GetSnapshot(), nil
}

func (s *Store) MergeNodeStates(sourceId, targetId types.ReplicaID) (crdt.GCounterSnapshot, error) {
	if _, exists := s.nodes[sourceId]; !exists {
		return crdt.GCounterSnapshot{}, fmt.Errorf("source node %d does not exist", sourceId)
	}

	if _, exists := s.nodes[targetId]; !exists {
		return crdt.GCounterSnapshot{}, fmt.Errorf("target node %d does not exist", targetId)
	}

	SourceCounterSnapshot := s.nodes[sourceId].Counter.GetSnapshot()

	s.nodes[targetId].Counter.Merge(&SourceCounterSnapshot)

	return s.nodes[targetId].Counter.GetSnapshot(), nil
}

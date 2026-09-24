package simulator

import (
	"fmt"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/node"
	"github.com/marver003/crdt/internal/types"
)

type Store struct {
	Nodes    map[types.ReplicaId]*node.Node
	CrdtType types.CrdtType
}

func NewStore(crdtType types.CrdtType) *Store {
	return &Store{
		Nodes:    make(map[types.ReplicaId]*node.Node),
		CrdtType: crdtType,
	}
}

func (s *Store) CreateNode(id types.ReplicaId) error {

	if _, exists := s.Nodes[id]; exists {
		return fmt.Errorf("node %q already exists", id)
	}

	n, err := node.CreateSimNode(id, s.CrdtType)
	if err != nil {
		return err
	}

	s.Nodes[id] = n

	return nil
}

func (s *Store) CreateNodeWithSnapshot(id types.ReplicaId, snapshot crdt.Snapshot) error {
	n, err := node.CreateSimNode(id, s.CrdtType)
	if err != nil {
		return err
	}

	if err := n.Crdt.ApplySnapshot(snapshot); err != nil {
		return err
	}

	s.Nodes[id] = n

	return nil
}

func (s *Store) RemoveNode(id types.ReplicaId) error {
	if _, exists := s.Nodes[id]; !exists {
		return fmt.Errorf("node %q does not exist", id)
	}

	delete(s.Nodes, id)

	return nil
}

// GetState returns the state of every gcounter node
func (s *Store) GetState() (map[types.ReplicaId]map[types.ReplicaId]uint64, error) {
	state := make(map[types.ReplicaId]map[types.ReplicaId]uint64, len(s.Nodes))

	for id, node := range s.Nodes {
		gcounter, ok := node.Crdt.(*crdt.GCounter)
		if !ok {
			return nil, fmt.Errorf("state is only supported for %q, simulator runs %q", types.GCounterType, s.CrdtType)
		}

		state[id] = gcounter.GetSnapshotCounts().Counts
	}

	return state, nil
}

// IncrementNodeCounter increments gcounter and returns its snapshot
func (s *Store) IncrementNodeCounter(id types.ReplicaId) (crdt.Snapshot, error) {
	n, exists := s.Nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %q does not exist", id)
	}

	if err := n.Increment(); err != nil {
		return nil, err
	}

	return n.Crdt.GetSnapshot(), nil
}

// MergeNodeStates merges source state into target state
func (s *Store) MergeNodeStates(sourceId, targetId types.ReplicaId) (crdt.Snapshot, error) {
	source, exists := s.Nodes[sourceId]
	if !exists {
		return nil, fmt.Errorf("source node %q does not exist", sourceId)
	}

	target, exists := s.Nodes[targetId]
	if !exists {
		return nil, fmt.Errorf("target node %q does not exist", targetId)
	}

	if source.Crdt.Type() != target.Crdt.Type() {
		return nil, fmt.Errorf("cannot merge %q into %q", source.Crdt.Type(), target.Crdt.Type())
	}

	target.Crdt.Merge(source.Crdt.GetSnapshot())

	return target.Crdt.GetSnapshot(), nil
}

package simulator

import (
	"fmt"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/node"
	"github.com/marver003/crdt/internal/types"
)

type Store struct {
	Nodes    map[types.ReplicaId]*node.Node
	Disabled map[types.ReplicaId]bool
	CrdtType types.CrdtType
}

func NewStore(crdtType types.CrdtType) *Store {
	return &Store{
		Nodes:    make(map[types.ReplicaId]*node.Node),
		Disabled: make(map[types.ReplicaId]bool),
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

func (s *Store) DisableNode(id types.ReplicaId) error {
	if _, exists := s.Nodes[id]; !exists {
		return fmt.Errorf("node %q does not exist", id)
	}

	s.Disabled[id] = true

	return nil
}

func (s *Store) EnableNode(id types.ReplicaId) error {
	if _, exists := s.Nodes[id]; !exists {
		return fmt.Errorf("node %q does not exist", id)
	}

	delete(s.Disabled, id)

	return nil
}

// IsUp reports whether the node exists and is not disabled
func (s *Store) IsUp(id types.ReplicaId) bool {
	_, exists := s.Nodes[id]

	return exists && !s.Disabled[id]
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

// GetGSetState returns the sorted elements of every gset node
func (s *Store) GetGSetState() (map[types.ReplicaId][]string, error) {
	state := make(map[types.ReplicaId][]string, len(s.Nodes))

	for id := range s.Nodes {
		elements, err := s.ReadNodeElements(id)
		if err != nil {
			return nil, err
		}

		state[id] = elements
	}

	return state, nil
}

// WriteNode applies a client write: an increment for gcounter, an add for gset
func (s *Store) WriteNode(id types.ReplicaId, element string) error {
	n, exists := s.Nodes[id]
	if !exists {
		return fmt.Errorf("node %q does not exist", id)
	}

	if s.CrdtType == types.GSetType {
		return n.Add(element)
	}

	return n.Increment()
}

// ReadNodeElements returns the sorted gset elements of a node
func (s *Store) ReadNodeElements(id types.ReplicaId) ([]string, error) {
	n, exists := s.Nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %q does not exist", id)
	}

	gset, ok := n.Crdt.(*crdt.GSet)
	if !ok {
		return nil, fmt.Errorf("elements are only supported for %q, simulator runs %q", types.GSetType, s.CrdtType)
	}

	return gset.Elements(), nil
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

// ReadNodeValue returns the gcounter value (sum of all counts) of a node
func (s *Store) ReadNodeValue(id types.ReplicaId) (uint64, error) {
	n, exists := s.Nodes[id]
	if !exists {
		return 0, fmt.Errorf("node %q does not exist", id)
	}

	gcounter, ok := n.Crdt.(*crdt.GCounter)
	if !ok {
		return 0, fmt.Errorf("read is only supported for %q, simulator runs %q", types.GCounterType, s.CrdtType)
	}

	return gcounter.Value(), nil
}

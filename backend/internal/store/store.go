package store

import (
	"fmt"
	"sync"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/types"
)

// Store is a thread-safe in-memory registry of named GCounter replicas.
type Store struct {
	mu       sync.RWMutex
	replicas map[types.ReplicaID]*crdt.GCounter
}

// New returns an empty Store.
func New() *Store {
	return &Store{
		replicas: make(map[types.ReplicaID]*crdt.GCounter),
	}
}

// Create adds a new replica with the given ID.
// Returns an error if a replica with that ID already exists.
func (s *Store) Create(id types.ReplicaID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.replicas[id]; exists {
		return fmt.Errorf("replica %q already exists", id)
	}

	s.replicas[id] = crdt.NewGCounter(id)
	return nil
}

// Get returns the GCounter for the given replica ID.
// Returns an error if no such replica exists.
func (s *Store) Get(id types.ReplicaID) (*crdt.GCounter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.replicas[id]
	if !ok {
		return nil, fmt.Errorf("replica %q not found", id)
	}

	return r, nil
}

// Delete removes the replica with the given ID.
// Returns an error if no such replica exists.
func (s *Store) Delete(id types.ReplicaID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.replicas[id]; !ok {
		return fmt.Errorf("replica %q not found", id)
	}

	delete(s.replicas, id)
	return nil
}

// ListSnapshot returns a snapshot of every replica's state, keyed by replica ID string.
func (s *Store) ListSnapshot() map[string]crdt.GCounterState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]crdt.GCounterState, len(s.replicas))
	for id, gc := range s.replicas {
		out[string(id)] = gc.State()
	}

	return out
}

// WithLock executes fn while holding the write lock on replica id.
func (s *Store) WithLock(id types.ReplicaID, fn func(*crdt.GCounter) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	r, ok := s.replicas[id]
	if !ok {
		return fmt.Errorf("replica %q not found", id)
	}

	return fn(r)
}

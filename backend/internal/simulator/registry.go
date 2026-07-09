package simulator

import (
	"fmt"

	"github.com/marver003/crdt/internal/types"
)

var id types.ReplicaID = 1

type NodeInfo struct {
	ID       types.ReplicaID
	HTTPAddr string
}

type Registry struct {
	nodes map[types.ReplicaID]NodeInfo
}

func NewRegistry() *Registry {
	return &Registry{
		make(map[types.ReplicaID]NodeInfo),
	}
}

func (r *Registry) Create() error {
	if _, exists := r.nodes[id]; exists {
		return fmt.Errorf("replica %d already exists", id)
	}

	r.nodes[id] = NodeInfo{ID: id, HTTPAddr: ":8081"}
	return nil
}

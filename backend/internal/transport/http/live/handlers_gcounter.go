package live

import (
	"log"
	"net/http"

	"github.com/marver003/crdt/internal/node"
	"github.com/marver003/crdt/internal/transport/http/helper"
	"github.com/marver003/crdt/internal/types"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	Node *node.Node
}

type replicaInfo struct {
	ID     types.ReplicaID            `json:"id"`
	Value  uint64                     `json:"value"`
	Counts map[types.ReplicaID]uint64 `json:"counts"`
}

// NewHandler returns a Handler backed by the given Node.
func NewHandler(n *node.Node) *Handler {
	log.Println("Creating new handler")
	return &Handler{Node: n}
}

// GetReplicaState handles GET /replicaState
func (h *Handler) GetReplicaState(w http.ResponseWriter, r *http.Request) {
	snapshot := h.Node.ListSnapshot()

	replInfo := replicaInfo{
		ID:     h.Node.ID,
		Value:  snapshot.Value,
		Counts: snapshot.Counts,
	}

	helper.WriteJSON(w, http.StatusOK, replInfo)
}

// IncrementReplica handles POST /increment
func (h *Handler) IncrementReplica(w http.ResponseWriter, r *http.Request) {
	h.Node.Increment()

	snapshot := h.Node.ListSnapshot()
	replInfo := replicaInfo{
		ID:     h.Node.ID,
		Value:  snapshot.Value,
		Counts: snapshot.Counts,
	}

	helper.WriteJSON(w, http.StatusOK, replInfo)
}

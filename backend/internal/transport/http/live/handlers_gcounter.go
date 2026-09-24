package live

import (
	"log"
	"net/http"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/node"
	"github.com/marver003/crdt/internal/transport/http/helper"
	"github.com/marver003/crdt/internal/types"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	Node *node.Node
}

type replicaInfo struct {
	Id     types.ReplicaId            `json:"id"`
	Value  uint64                     `json:"value"`
	Counts map[types.ReplicaId]uint64 `json:"counts"`
}

// NewHandler returns a Handler backed by the given Node.
func NewHandler(n *node.Node) *Handler {
	log.Println("Creating new handler")
	return &Handler{Node: n}
}

// GetReplicaState handles GET /replicaState
func (h *Handler) GetReplicaState(w http.ResponseWriter, r *http.Request) {
	h.writeReplicaInfo(w)
}

// IncrementReplica handles POST /increment
func (h *Handler) IncrementReplica(w http.ResponseWriter, r *http.Request) {
	if err := h.Node.Increment(); err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeReplicaInfo(w)
}

func (h *Handler) writeReplicaInfo(w http.ResponseWriter) {
	snapshot, ok := h.Node.ListSnapshot().(*crdt.GCounterSnapshot)
	if !ok {
		helper.WriteError(w, http.StatusInternalServerError, "node is not a gcounter")
		return
	}

	replInfo := replicaInfo{
		Id:     h.Node.Id,
		Value:  snapshot.Value,
		Counts: snapshot.Counts,
	}

	helper.WriteJSON(w, http.StatusOK, replInfo)
}

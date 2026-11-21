package live

import (
	"net/http"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/transport/http/helper"
	"github.com/marver003/crdt/internal/types"
)

type counterInfo struct {
	Id     types.ReplicaId            `json:"id"`
	Value  uint64                     `json:"value"`
	Counts map[types.ReplicaId]uint64 `json:"counts"`
}

// GetCounterState handles GET /replicaState
func (h *Handler) GetCounterState(w http.ResponseWriter, r *http.Request) {
	h.writeCounterInfo(w)
}

// IncrementCounter handles POST /increment
func (h *Handler) IncrementCounter(w http.ResponseWriter, r *http.Request) {
	if err := h.Node.Increment(); err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeCounterInfo(w)
}

func (h *Handler) writeCounterInfo(w http.ResponseWriter) {
	snapshot, ok := h.Node.ListSnapshot().(*crdt.GCounterSnapshot)
	if !ok {
		helper.WriteError(w, http.StatusInternalServerError, "node is not a gcounter")
		return
	}

	helper.WriteJSON(w, http.StatusOK, counterInfo{
		Id:     h.Node.Id,
		Value:  snapshot.Value,
		Counts: snapshot.Counts,
	})
}

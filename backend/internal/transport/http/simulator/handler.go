package internal

import (
	"encoding/json"
	"net/http"

	"github.com/marver003/crdt/internal/simulator"
	httptransport "github.com/marver003/crdt/internal/transport/http"
	"github.com/marver003/crdt/internal/types"
)

type Handler struct {
	Registry *simulator.Registry
}

// NewHandler returns a Handler backed by the given Registry.
func NewHandler(r *simulator.Registry) *Handler {
	return &Handler{Registry: r}
}

// ------------------------------------------------------------------ POST /replicas

type createRequest struct {
	ID types.ReplicaID `json:"id"`
}

type createResponse struct {
	ID    types.ReplicaID `json:"id"`
	Value uint64          `json:"value"`
}

// CreateReplica handles POST /replicas
// Body: { "id": "replica-A" }
func (h *Handler) CreateReplica(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httptransport.WriteError(w, http.StatusBadRequest, "request body must be JSON with a non-empty \"id\" field")
		return
	}

	if err := h.Registry.Create(); err != nil {
		httptransport.WriteError(w, http.StatusConflict, err.Error())
		return
	}

	httptransport.WriteJSON(w, http.StatusCreated, createResponse{ID: req.ID, Value: 0})
}

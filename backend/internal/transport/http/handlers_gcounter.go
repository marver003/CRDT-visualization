package internal

import (
	"encoding/json"
	"net/http"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/store"
	"github.com/marver003/crdt/internal/types"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	Store *store.Store
}

// NewHandler returns a Handler backed by the given Store.
func NewHandler(s *store.Store) *Handler {
	return &Handler{Store: s}
}

// ------------------------------------------------------------------ POST /replicas

type createRequest struct {
	ID string `json:"id"`
}

type createResponse struct {
	ID    string `json:"id"`
	Value uint64 `json:"value"`
}

// CreateReplica handles POST /replicas
// Body: { "id": "replica-A" }
func (h *Handler) CreateReplica(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		writeError(w, http.StatusBadRequest, "request body must be JSON with a non-empty \"id\" field")
		return
	}

	id := types.ReplicaID(req.ID)
	if err := h.Store.Create(id); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, createResponse{ID: req.ID, Value: 0})
}

// ------------------------------------------------------------------ GET /replicas

type replicaInfo struct {
	ID     string            `json:"id"`
	Value  uint64            `json:"value"`
	Counts map[string]uint64 `json:"counts"`
}

// ListReplicas handles GET /replicas
func (h *Handler) ListReplicas(w http.ResponseWriter, r *http.Request) {
	snapshots := h.Store.ListSnapshot()

	list := make([]replicaInfo, 0, len(snapshots))
	for id, state := range snapshots {
		counts := make(map[string]uint64, len(state.Counts))
		var total uint64
		for rid, v := range state.Counts {
			counts[string(rid)] = v
			total += v
		}
		list = append(list, replicaInfo{ID: id, Value: total, Counts: counts})
	}

	writeJSON(w, http.StatusOK, list)
}

// ------------------------------------------------------------------ GET /replicas/{id}

// GetReplica handles GET /replicas/{id}
func (h *Handler) GetReplica(w http.ResponseWriter, r *http.Request) {
	id := types.ReplicaID(r.PathValue("id"))

	gc, err := h.Store.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	state := gc.State()
	counts := make(map[string]uint64, len(state.Counts))
	var total uint64
	for rid, v := range state.Counts {
		counts[string(rid)] = v
		total += v
	}

	writeJSON(w, http.StatusOK, replicaInfo{ID: string(id), Value: total, Counts: counts})
}

// ------------------------------------------------------------------ POST /replicas/{id}/increment

// IncrementReplica handles POST /replicas/{id}/increment
func (h *Handler) IncrementReplica(w http.ResponseWriter, r *http.Request) {
	id := types.ReplicaID(r.PathValue("id"))

	var info replicaInfo
	err := h.Store.WithLock(id, func(gc *crdt.GCounter) error {
		gc.Increment()

		state := gc.State()
		counts := make(map[string]uint64, len(state.Counts))
		var total uint64
		for rid, v := range state.Counts {
			counts[string(rid)] = v
			total += v
		}
		info = replicaInfo{ID: string(id), Value: total, Counts: counts}
		return nil
	})

	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, info)
}

// ------------------------------------------------------------------ POST /replicas/{id}/add

type addRequest struct {
	N uint64 `json:"n"`
}

// AddToReplica handles POST /replicas/{id}/add
// Body: { "n": 5 }
func (h *Handler) AddToReplica(w http.ResponseWriter, r *http.Request) {
	id := types.ReplicaID(r.PathValue("id"))

	var req addRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "request body must be JSON with an \"n\" field")
		return
	}

	var info replicaInfo
	err := h.Store.WithLock(id, func(gc *crdt.GCounter) error {
		gc.Add(req.N)

		state := gc.State()
		counts := make(map[string]uint64, len(state.Counts))
		var total uint64
		for rid, v := range state.Counts {
			counts[string(rid)] = v
			total += v
		}
		info = replicaInfo{ID: string(id), Value: total, Counts: counts}
		return nil
	})

	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, info)
}

// ------------------------------------------------------------------ POST /replicas/{id}/merge

type mergeRequest struct {
	// Counts is the GCounterState.Counts map from the other replica.
	Counts map[string]uint64 `json:"counts"`
}

// MergeReplica handles POST /replicas/{id}/merge
// Body: { "counts": { "replica-B": 3, "replica-C": 7 } }
func (h *Handler) MergeReplica(w http.ResponseWriter, r *http.Request) {
	id := types.ReplicaID(r.PathValue("id"))

	var req mergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Counts == nil {
		writeError(w, http.StatusBadRequest, "request body must be JSON with a \"counts\" map")
		return
	}

	// Convert string keys to ReplicaID keys.
	incoming := make(map[types.ReplicaID]uint64, len(req.Counts))
	for k, v := range req.Counts {
		incoming[types.ReplicaID(k)] = v
	}

	var info replicaInfo
	err := h.Store.WithLock(id, func(gc *crdt.GCounter) error {
		gc.ApplyState(crdt.GCounterState{Counts: incoming})

		state := gc.State()
		counts := make(map[string]uint64, len(state.Counts))
		var total uint64
		for rid, v := range state.Counts {
			counts[string(rid)] = v
			total += v
		}
		info = replicaInfo{ID: string(id), Value: total, Counts: counts}
		return nil
	})

	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, info)
}

// ------------------------------------------------------------------ DELETE /replicas/{id}

// DeleteReplica handles DELETE /replicas/{id}
func (h *Handler) DeleteReplica(w http.ResponseWriter, r *http.Request) {
	id := types.ReplicaID(r.PathValue("id"))

	if err := h.Store.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

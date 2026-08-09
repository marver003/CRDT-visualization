package simulator

import (
	"encoding/json"
	"net/http"

	"github.com/marver003/crdt/internal/transport/http/helper"
	"github.com/marver003/crdt/internal/types"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	Store *Store
}

func NewHandler(s *Store) *Handler {
	return &Handler{Store: s}
}

func (h *Handler) NotImplemented(w http.ResponseWriter, r *http.Request) {
	helper.WriteError(w, http.StatusNotImplemented, "Not yet implemented.")
}

type reqCreateNode struct {
	ID types.ReplicaID `json:"id"`
}

type respCreateNode struct {
	ID    types.ReplicaID `json:"id"`
	Value uint64          `json:"value"`
}

func (h *Handler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var req reqCreateNode

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "id" field`)
		return
	}

	if err = h.Store.CreateNode(req.ID); err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.WriteJSON(w, http.StatusCreated, &respCreateNode{ID: req.ID, Value: 0})
}

func (h *Handler) RemoveNode(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ParseReplicaId(r.PathValue("replicaId"))

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = h.Store.RemoveNode(id); err != nil {
		helper.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	helper.WriteJSON(w, http.StatusOK, &struct{}{})
}

type respGetState struct {
	State map[types.ReplicaID]map[types.ReplicaID]uint64 `json:"state"`
}

func (h *Handler) GetState(w http.ResponseWriter, r *http.Request) {
	state := h.Store.GetState()

	helper.WriteJSON(w, http.StatusOK, &respGetState{State: state})
}

type reqIncrementNode struct {
	ID types.ReplicaID `json:"id"`
}

type respIncrementNode struct {
	Counts map[types.ReplicaID]uint64
	Value  uint64
}

func (h *Handler) IncrementNode(w http.ResponseWriter, r *http.Request) {
	var req reqIncrementNode

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "id" field`)
		return
	}

	snapshot, err := h.Store.IncrementNodeCounter(req.ID)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.WriteJSON(w, http.StatusOK, &respIncrementNode{Counts: snapshot.Counts, Value: snapshot.Value})
}

type reqMergeNodes struct {
	SourceId types.ReplicaID `json:"source-id"`
	TargetId types.ReplicaID `json:"target-id"`
}

type respMergeNodes struct {
	Counts map[types.ReplicaID]uint64
	Value  uint64
}

func (h *Handler) MergeNodes(w http.ResponseWriter, r *http.Request) {
	var req reqMergeNodes

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "source-id" and "target-id fields`)
		return
	}

	snapshot, err := h.Store.MergeNodeStates(req.SourceId, req.TargetId)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.WriteJSON(w, http.StatusOK, &respMergeNodes{Counts: snapshot.Counts, Value: snapshot.Value})
}

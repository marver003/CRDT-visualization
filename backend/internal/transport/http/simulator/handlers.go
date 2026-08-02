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
	id, err := helper.ParseReplicaId(r.PathValue("id"))

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

func (h *Handler) IncrementNode(w http.ResponseWriter, r *http.Request) {
	helper.WriteError(w, http.StatusNotImplemented, "Not yet implemented.")
}

func (h *Handler) MergeNodes(w http.ResponseWriter, r *http.Request) {
	helper.WriteError(w, http.StatusNotImplemented, "Not yet implemented.")
}

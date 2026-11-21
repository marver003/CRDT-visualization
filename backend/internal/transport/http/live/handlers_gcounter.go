package live

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

const maxCounterAddAmount uint64 = 1000

type addAmountRequest struct {
	Amount *uint64 `json:"amount"`
}

// GetCounterState handles GET /replicaState
func (h *Handler) GetCounterState(w http.ResponseWriter, r *http.Request) {
	h.writeCounterInfo(w)
}

// AddAmountCounter handles POST /add for a gcounter node
func (h *Handler) AddAmountCounter(w http.ResponseWriter, r *http.Request) {
	var request addAmountRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil && !errors.Is(err, io.EOF) {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON body, amount must be a positive whole number")
		return
	}

	amount := uint64(1)

	if request.Amount != nil {
		amount = *request.Amount
	}

	if amount == 0 || amount > maxCounterAddAmount {
		helper.WriteError(w, http.StatusBadRequest, fmt.Sprintf("amount must be between 1 and %d", maxCounterAddAmount))
		return
	}

	if err := h.Node.IncrementBy(amount); err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.WriteOkNoContent(w)
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

package live

import (
	"encoding/json"
	"maps"
	"net/http"
	"slices"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/transport/http/helper"
	"github.com/marver003/crdt/internal/types"
)

type setInfo struct {
	Id       types.ReplicaId `json:"id"`
	Elements []string        `json:"elements"`
}

type addRequest struct {
	Element string `json:"element"`
}

// GetSetState handles GET /replicaState for a gset node
func (h *Handler) GetSetState(w http.ResponseWriter, r *http.Request) {
	h.writeSetInfo(w)
}

// AddElement handles POST /add for a gset node
func (h *Handler) AddElement(w http.ResponseWriter, r *http.Request) {
	var request addRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if request.Element == "" {
		helper.WriteError(w, http.StatusBadRequest, "element must not be empty")
		return
	}

	if err := h.Node.Add(request.Element); err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.Node.GossipService.Notify()

	helper.WriteOkNoContent(w)
}

func (h *Handler) writeSetInfo(w http.ResponseWriter) {
	snapshot, ok := h.Node.ListSnapshot().(*crdt.GSetSnapshot)
	if !ok {
		helper.WriteError(w, http.StatusInternalServerError, "node is not a gset")
		return
	}

	helper.WriteJSON(w, http.StatusOK, setInfo{
		Id:       h.Node.Id,
		Elements: slices.Sorted(maps.Keys(snapshot.Set)),
	})
}

package simhttp

import (
	"encoding/json"
	"net/http"

	"github.com/marver003/crdt/internal/simulator"
	"github.com/marver003/crdt/internal/transport/http/helper"
	"github.com/marver003/crdt/internal/types"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	Sim *simulator.Simulator
}

func NewHandler(s *simulator.Simulator) *Handler {
	return &Handler{Sim: s}
}

type reqCreateNode struct {
	ID types.ReplicaID `json:"id"`
}

func (h *Handler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var req reqCreateNode

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "id" field`)
		return
	}

	h.Sim.Queue.Enqueue(
		simulator.CreateNodeOperation{
			StepID:    h.Sim.GetNextStepId(),
			ReplicaID: req.ID,
		})

	helper.WriteOkNoContent(w)
}

func (h *Handler) RemoveNode(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ParseReplicaId(r.PathValue("replicaId"))

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.Sim.Queue.Enqueue(
		simulator.RemoveNodeOperation{
			StepID:    h.Sim.GetNextStepId(),
			ReplicaID: id,
		})

	helper.WriteOkNoContent(w)
}

type reqIncrementNode struct {
	ID types.ReplicaID `json:"id"`
}

func (h *Handler) IncrementNode(w http.ResponseWriter, r *http.Request) {
	var req reqIncrementNode

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "id" field`)
		return
	}

	h.Sim.Queue.Enqueue(
		simulator.IncrementOperation{
			StepID:    h.Sim.GetNextStepId(),
			ReplicaID: req.ID,
		})

	helper.WriteOkNoContent(w)
}

type reqMergeNodes struct {
	SourceId types.ReplicaID `json:"source-id"`
	TargetId types.ReplicaID `json:"target-id"`
}

func (h *Handler) MergeNodes(w http.ResponseWriter, r *http.Request) {
	var req reqMergeNodes

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "source-id" and "target-id fields`)
		return
	}

	h.Sim.Queue.Enqueue(
		simulator.SendGossipMessageOperation{
			StepID: h.Sim.GetNextStepId(),
			From:   req.SourceId,
			To:     req.TargetId,
		})

	helper.WriteOkNoContent(w)
}

type respGetState struct {
	State map[types.ReplicaID]map[types.ReplicaID]uint64 `json:"state"`
}

func (h *Handler) GetState(w http.ResponseWriter, r *http.Request) {
	state := h.Sim.Store.GetState()

	helper.WriteJSON(w, http.StatusOK, &respGetState{State: state})
}

type respGetSteps struct {
	Steps []respStep `json:"steps"`
}

type respStep struct {
	StepID types.StepID `json:"stepId"`
	Type   string       `json:"type"`

	ReplicaID types.ReplicaID `json:"replicaId,omitempty"`

	From types.ReplicaID `json:"from,omitempty"`
	To   types.ReplicaID `json:"to,omitempty"`
}

func (h *Handler) GetSteps(w http.ResponseWriter, r *http.Request) {
	operations := h.Sim.Queue.Operations

	steps := make([]respStep, 0, len(operations))

	for _, op := range operations {
		switch op := op.(type) {
		case simulator.CreateNodeOperation:
			steps = append(steps, respStep{
				StepID:    op.StepID,
				Type:      "create_node",
				ReplicaID: op.ReplicaID,
			})

		case simulator.RemoveNodeOperation:
			steps = append(steps, respStep{
				StepID:    op.StepID,
				Type:      "remove_node",
				ReplicaID: op.ReplicaID,
			})

		case simulator.IncrementOperation:
			steps = append(steps, respStep{
				StepID:    op.StepID,
				Type:      "increment",
				ReplicaID: op.ReplicaID,
			})

		case simulator.SendGossipMessageOperation:
			steps = append(steps, respStep{
				StepID: op.StepID,
				Type:   "send_gossip",
				From:   op.From,
				To:     op.To,
			})

		case simulator.ReceiveGossipMessageOperation:
			steps = append(steps, respStep{
				StepID: op.StepID,
				Type:   "receive_gossip",
				From:   op.Message.From,
				To:     op.Message.To,
			})
		}
	}

	helper.WriteJSON(w, http.StatusOK, &respGetSteps{
		Steps: steps,
	})
}

func (h *Handler) NextStep(w http.ResponseWriter, r *http.Request) {
	h.Sim.Step()

	helper.WriteOkNoContent(w)
}

func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	h.Sim = simulator.New()

	helper.WriteOkNoContent(w)
}

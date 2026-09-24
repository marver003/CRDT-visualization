package simhttp

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/marver003/crdt/internal/crdt"
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

func (h *Handler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var req reqCreateNode

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "id" field`)
		return
	}

	stepId := h.Sim.GetNextStepId()

	h.Sim.Queue.Enqueue(
		simulator.CreateNodeOperation{
			BaseOperation: simulator.BaseOperation{StepId: stepId},
			ReplicaId:     req.Id,
		})

	helper.WriteOkNoContent(w)
}

func (h *Handler) RemoveNode(w http.ResponseWriter, r *http.Request) {
	id := types.ReplicaId(r.PathValue("replicaId"))

	stepId := h.Sim.GetNextStepId()

	h.Sim.Queue.Enqueue(
		simulator.RemoveNodeOperation{
			BaseOperation: simulator.BaseOperation{StepId: stepId},
			ReplicaId:     id,
		})

	helper.WriteOkNoContent(w)
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
			BaseOperation: simulator.BaseOperation{StepId: h.Sim.GetNextStepId()},
			ReplicaId:     req.Id,
		})

	stepIdSend := h.Sim.GetNextStepId()
	stepIdReceive := h.Sim.GetNextStepId()

	for _, node := range h.Sim.Store.Nodes {
		if req.Id == node.Id {
			continue
		}

		messageId := h.Sim.NewMessage(req.Id, node.Id, nil)
		message := h.Sim.PendingMessages[messageId]

		h.Sim.Queue.Enqueue(simulator.SendGossipMessageOperation{
			BaseOperation: simulator.BaseOperation{StepId: stepIdSend},
			From:          req.Id,
			To:            node.Id,
			Message:       message,
		})

		h.Sim.Queue.Enqueue(simulator.ReceiveGossipMessageOperation{
			BaseOperation: simulator.BaseOperation{StepId: stepIdReceive},
			Message:       message,
		})
	}

	helper.WriteOkNoContent(w)
}

func (h *Handler) MergeNodes(w http.ResponseWriter, r *http.Request) {
	var req reqMergeNodes

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "source-id" and "target-id fields`)
		return
	}

	messageId := h.Sim.NewMessage(req.SourceId, req.TargetId, nil)
	message := h.Sim.PendingMessages[messageId]

	stepId := h.Sim.GetNextStepId()
	h.Sim.Queue.Enqueue(
		simulator.SendGossipMessageOperation{
			BaseOperation: simulator.BaseOperation{StepId: stepId},
			From:          req.SourceId,
			To:            req.TargetId,
			Message:       message,
		})

	stepId = h.Sim.GetNextStepId()
	h.Sim.Queue.Enqueue(simulator.ReceiveGossipMessageOperation{
		BaseOperation: simulator.BaseOperation{StepId: stepId},
		Message:       message,
	})

	helper.WriteOkNoContent(w)
}

func (h *Handler) GetState(w http.ResponseWriter, r *http.Request) {
	state, err := h.Sim.Store.GetState()
	if err != nil {
		helper.WriteError(w, http.StatusNotImplemented, err.Error())
		return
	}

	helper.WriteJSON(w, http.StatusOK, &respGetState{State: state})
}

func (h *Handler) GetSteps(w http.ResponseWriter, r *http.Request) {
	operations := h.Sim.Queue.Operations

	steps := make([]respStep, 0, len(operations))

	for _, ops := range operations {
		for _, op := range ops {
			switch op := op.(type) {
			case simulator.CreateNodeOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      "create_node",
					ReplicaId: op.ReplicaId,
				})

			case simulator.RemoveNodeOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      "remove_node",
					ReplicaId: op.ReplicaId,
				})

			case simulator.IncrementOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      "increment",
					ReplicaId: op.ReplicaId,
				})

			case simulator.SendGossipMessageOperation:
				steps = append(steps, respStep{
					StepId: op.StepId,
					Type:   "send_gossip",
					From:   op.From,
					To:     op.To,
				})

			case simulator.ReceiveGossipMessageOperation:
				steps = append(steps, respStep{
					StepId: op.StepId,
					Type:   "receive_gossip",
					From:   op.Message.From,
					To:     op.Message.To,
				})
			}
		}

	}

	// sort steps ASC
	slices.SortFunc(steps, func(a, b respStep) int {
		return int(a.StepId) - int(b.StepId)
	})

	helper.WriteJSON(w, http.StatusOK, &respGetSteps{
		Steps: steps,
	})
}

func (h *Handler) NextStep(w http.ResponseWriter, r *http.Request) {
	if h.Sim.ExecutingStepId == h.Sim.NextStepId {
		helper.WriteError(w, http.StatusBadRequest, "There is no next step")
		return
	}

	h.Sim.Step()

	helper.WriteOkNoContent(w)
}

func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	h.Sim = simulator.New(h.Sim.Store.CrdtType)

	helper.WriteOkNoContent(w)
}

func (h *Handler) Load(w http.ResponseWriter, r *http.Request) {

	var req reqLoad

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	snapshots := make(map[types.ReplicaId]crdt.Snapshot)

	for nodeId, state := range req.State {
		snapshots[types.ReplicaId(nodeId)] = &crdt.GCounterSnapshot{
			Counts: state,
			Value:  0, // value is not important for applying the snapshot
		}
	}

	sim, err := simulator.Load(types.GCounterType, snapshots)
	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.Sim = sim

	helper.WriteOkNoContent(w)
}

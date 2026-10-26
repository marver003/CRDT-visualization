package simhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/marver003/crdt/internal/simulator"
	"github.com/marver003/crdt/internal/transport/http/helper"
	"github.com/marver003/crdt/internal/types"
)

// NewSimFunc builds a fresh simulator of the given CRDT type
type NewSimFunc func(crdtType types.CrdtType) (*simulator.Simulator, error)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	mu     sync.Mutex
	Sim    *simulator.Simulator
	newSim NewSimFunc
}

func (h *Handler) locked(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		defer h.mu.Unlock()

		next(w, r)
	}
}

func NewHandler(s *simulator.Simulator, newSim NewSimFunc) *Handler {
	return &Handler{Sim: s, newSim: newSim}
}

func (h *Handler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var req reqCreateNode

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil || req.Id == "" {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "id" field`)
		return
	}

	if err := h.Sim.Store.CreateNode(req.Id); err != nil {
		helper.WriteError(w, http.StatusConflict, err.Error())
		return
	}

	helper.WriteOkNoContent(w)
}

func (h *Handler) DisableNode(w http.ResponseWriter, r *http.Request) {
	id := types.ReplicaId(r.PathValue("replicaId"))

	if err := h.Sim.Store.DisableNode(id); err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	helper.WriteOkNoContent(w)
}

func (h *Handler) EnableNode(w http.ResponseWriter, r *http.Request) {
	id := types.ReplicaId(r.PathValue("replicaId"))

	wasDisabled := h.Sim.Store.Disabled[id]

	if err := h.Sim.Store.EnableNode(id); err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if wasDisabled {
		h.Sim.OnNodeEnabled(id)
	}

	helper.WriteOkNoContent(w)
}

func (h *Handler) WriteRequest(w http.ResponseWriter, r *http.Request) {
	h.clientRequest(w, r, simulator.WriteRequest)
}

func (h *Handler) ReadRequest(w http.ResponseWriter, r *http.Request) {
	h.clientRequest(w, r, simulator.ReadRequest)
}

func (h *Handler) clientRequest(w http.ResponseWriter, r *http.Request, kind simulator.RequestKind) {
	var req reqClientRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "clientId" and "nodeId" fields`)
		return
	}

	if _, ok := h.Sim.Clients[req.ClientId]; !ok {
		helper.WriteError(w, http.StatusBadRequest, "Client does not exist")
		return
	}

	if _, ok := h.Sim.Store.Nodes[req.NodeId]; !ok {
		helper.WriteError(w, http.StatusBadRequest, "Node does not exist")
		return
	}

	if err := validateGSetWrite(h.Sim.Store.CrdtType, kind, req.Element); err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	amount, err := gcounterAmount(h.Sim.Store.CrdtType, kind, req.Amount)
	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Sim.SendClientRequest(req.ClientId, req.NodeId, kind, req.Element, amount); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, simulator.ErrClientBusy) {
			status = http.StatusConflict
		}

		helper.WriteError(w, status, err.Error())
		return
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

	h.Sim.Queue.Enqueue(
		simulator.SendGossipMessageOperation{
			BaseOperation: h.Sim.After(1),
			From:          req.SourceId,
			To:            req.TargetId,
		})

	helper.WriteOkNoContent(w)
}

func (h *Handler) GetState(w http.ResponseWriter, r *http.Request) {
	state, err := stateFor(h.Sim.Store)
	if err != nil {
		helper.WriteError(w, http.StatusNotImplemented, err.Error())
		return
	}

	resp := &respGetState{
		Crdt:     h.Sim.Store.CrdtType,
		Now:      h.Sim.Now,
		State:    state,
		Disabled: disabledNodes(h.Sim.Store),
		Clients:  make([]respClient, 0, len(h.Sim.Clients)),
		InFlight: make([]respMessage, 0, len(h.Sim.PendingMessages)),
	}

	for _, client := range h.Sim.Clients {
		item := respClient{Id: client.Id, Locked: client.Locked(h.Sim.Now)}

		if client.LastResult != nil {
			item.Result = &respClientResult{
				Kind:     string(client.LastResult.Kind),
				NodeId:   client.LastResult.ReplicaId,
				Value:    client.LastResult.Value,
				Elements: client.LastResult.Elements,
				At:       client.LastResult.At,
			}
		}

		resp.Clients = append(resp.Clients, item)
	}

	slices.SortFunc(resp.Clients, func(a, b respClient) int {
		return strings.Compare(string(a.Id), string(b.Id))
	})

	for _, msg := range h.Sim.PendingMessages {
		resp.InFlight = append(resp.InFlight, respMessage{
			Id:        msg.Id,
			Kind:      string(msg.Kind),
			From:      respEndpoint{Kind: string(msg.From.Kind), Id: msg.From.Id},
			To:        respEndpoint{Kind: string(msg.To.Kind), Id: msg.To.Id},
			SentAt:    msg.SentAt,
			DeliverAt: msg.DeliverAt,
			ClientId:  msg.ClientId,
			Seq:       msg.Seq,
		})
	}

	slices.SortFunc(resp.InFlight, func(a, b respMessage) int {
		if a.SentAt != b.SentAt {
			return int(a.SentAt) - int(b.SentAt)
		}
		return strings.Compare(a.Id, b.Id)
	})

	helper.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetSteps(w http.ResponseWriter, r *http.Request) {
	operations := h.Sim.Queue.Operations

	steps := make([]respStep, 0, len(operations))

	for _, ops := range operations {
		for _, op := range ops {
			switch op := op.(type) {
			case simulator.ReceiveRequestOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      "receive_" + string(op.Message.Kind),
					ClientId:  types.ClientId(op.Message.From.Id),
					ReplicaId: types.ReplicaId(op.Message.To.Id),
					Seq:       op.Message.Seq,
				})

			case simulator.IncrementOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      writeStepType(h.Sim.Store.CrdtType),
					ReplicaId: op.ReplicaId,
					ClientId:  op.ClientId,
					Seq:       op.Seq,
					Element:   op.Element,
				})

			case simulator.ReadOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      "read",
					ReplicaId: op.ReplicaId,
					ClientId:  op.ClientId,
					Seq:       op.Seq,
				})

			case simulator.SendResponseOperation:
				stepType := "send_ack"
				if op.Kind == simulator.ReadResponseMessage {
					stepType = "send_response"
				}

				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      stepType,
					ReplicaId: op.ReplicaId,
					ClientId:  op.ClientId,
					Seq:       op.Seq,
				})

			case simulator.ReceiveResponseOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      "receive_" + string(op.Message.Kind),
					ReplicaId: types.ReplicaId(op.Message.From.Id),
					ClientId:  types.ClientId(op.Message.To.Id),
					Seq:       op.Message.Seq,
				})

			case simulator.SendGossipMessageOperation:
				steps = append(steps, respStep{
					StepId:   op.StepId,
					Type:     "send_gossip",
					From:     op.From,
					To:       op.To,
					ClientId: op.ClientId,
					Seq:      op.Seq,
					Replay:   op.Replay,
				})

			case simulator.ReceiveGossipMessageOperation:
				steps = append(steps, respStep{
					StepId:   op.StepId,
					Type:     "receive_gossip",
					From:     types.ReplicaId(op.Message.From.Id),
					To:       types.ReplicaId(op.Message.To.Id),
					ClientId: op.Message.ClientId,
					Seq:      op.Message.Seq,
					Replay:   op.Message.Replay,
				})

			case simulator.SendGossipAckOperation:
				steps = append(steps, respStep{
					StepId:   op.StepId,
					Type:     "send_gossip_ack",
					From:     op.From,
					To:       op.To,
					ClientId: op.ClientId,
					Seq:      op.Seq,
					Replay:   op.Replay,
				})

			case simulator.ReceiveGossipAckOperation:
				steps = append(steps, respStep{
					StepId:   op.StepId,
					Type:     "receive_gossip_ack",
					From:     types.ReplicaId(op.Message.From.Id),
					To:       types.ReplicaId(op.Message.To.Id),
					ClientId: op.Message.ClientId,
					Seq:      op.Message.Seq,
					Replay:   op.Message.Replay,
				})

			case simulator.SendSyncRequestOperation:
				steps = append(steps, respStep{
					StepId: op.StepId,
					Type:   "send_sync_request",
					From:   op.From,
					To:     op.To,
				})

			case simulator.ReceiveSyncRequestOperation:
				steps = append(steps, respStep{
					StepId: op.StepId,
					Type:   "receive_sync_request",
					From:   types.ReplicaId(op.Message.From.Id),
					To:     types.ReplicaId(op.Message.To.Id),
				})

			case simulator.MergeOperation:
				steps = append(steps, respStep{
					StepId:   op.StepId,
					Type:     "merge",
					From:     types.ReplicaId(op.Message.From.Id),
					To:       types.ReplicaId(op.Message.To.Id),
					ClientId: op.Message.ClientId,
					Seq:      op.Message.Seq,
					Replay:   op.Message.Replay,
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
	if h.Sim.Queue.Empty() {
		helper.WriteError(w, http.StatusBadRequest, "There is no next step")
		return
	}

	h.Sim.Step()

	helper.WriteOkNoContent(w)
}

func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	if h.newSim == nil {
		h.Sim = simulator.New(h.Sim.Store.CrdtType)

		helper.WriteOkNoContent(w)
		return
	}

	sim, err := h.newSim(h.Sim.Store.CrdtType)
	if err != nil {
		helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.Sim = sim

	helper.WriteOkNoContent(w)
}

func (h *Handler) SwitchSimulator(w http.ResponseWriter, r *http.Request) {
	var req reqSwitchSimulator

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, `Request body must be JSON with "crdt" field`)
		return
	}

	if req.Crdt != types.GCounterType && req.Crdt != types.GSetType && req.Crdt != types.GCounterOpType {
		helper.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported crdt %q", req.Crdt))
		return
	}

	h.Sim = simulator.New(req.Crdt)

	helper.WriteOkNoContent(w)
}

func disabledNodes(store *simulator.Store) []types.ReplicaId {
	disabled := make([]types.ReplicaId, 0, len(store.Disabled))

	for id := range store.Disabled {
		disabled = append(disabled, id)
	}

	slices.Sort(disabled)

	return disabled
}

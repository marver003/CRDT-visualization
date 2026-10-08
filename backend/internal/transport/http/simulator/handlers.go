package simhttp

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/marver003/crdt/internal/simulator"
	"github.com/marver003/crdt/internal/transport/http/helper"
	"github.com/marver003/crdt/internal/types"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	mu  sync.Mutex
	Sim *simulator.Simulator
}

func (h *Handler) locked(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		defer h.mu.Unlock()

		next(w, r)
	}
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

	h.Sim.Queue.Enqueue(
		simulator.CreateNodeOperation{
			BaseOperation: h.Sim.After(1),
			ReplicaId:     req.Id,
		})

	helper.WriteOkNoContent(w)
}

func (h *Handler) RemoveNode(w http.ResponseWriter, r *http.Request) {
	id := types.ReplicaId(r.PathValue("replicaId"))

	h.Sim.Queue.Enqueue(
		simulator.RemoveNodeOperation{
			BaseOperation: h.Sim.After(1),
			ReplicaId:     id,
		})

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

	h.Sim.SendClientRequest(req.ClientId, req.NodeId, kind)

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
	state, err := h.Sim.Store.GetState()
	if err != nil {
		helper.WriteError(w, http.StatusNotImplemented, err.Error())
		return
	}

	resp := &respGetState{
		Now:      h.Sim.Now,
		State:    state,
		Clients:  make([]respClient, 0, len(h.Sim.Clients)),
		InFlight: make([]respMessage, 0, len(h.Sim.PendingMessages)),
	}

	for _, client := range h.Sim.Clients {
		item := respClient{Id: client.Id}

		if client.LastResult != nil {
			item.Result = &respClientResult{
				Kind:   string(client.LastResult.Kind),
				NodeId: client.LastResult.ReplicaId,
				Value:  client.LastResult.Value,
				At:     client.LastResult.At,
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

			case simulator.ReceiveRequestOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      "receive_" + string(op.Message.Kind),
					ClientId:  types.ClientId(op.Message.From.Id),
					ReplicaId: types.ReplicaId(op.Message.To.Id),
				})

			case simulator.IncrementOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      "increment",
					ReplicaId: op.ReplicaId,
					ClientId:  op.ClientId,
				})

			case simulator.ReadOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      "read",
					ReplicaId: op.ReplicaId,
					ClientId:  op.ClientId,
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
				})

			case simulator.ReceiveResponseOperation:
				steps = append(steps, respStep{
					StepId:    op.StepId,
					Type:      "receive_" + string(op.Message.Kind),
					ReplicaId: types.ReplicaId(op.Message.From.Id),
					ClientId:  types.ClientId(op.Message.To.Id),
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
					From:   types.ReplicaId(op.Message.From.Id),
					To:     types.ReplicaId(op.Message.To.Id),
				})

			case simulator.MergeOperation:
				steps = append(steps, respStep{
					StepId: op.StepId,
					Type:   "merge",
					From:   types.ReplicaId(op.Message.From.Id),
					To:     types.ReplicaId(op.Message.To.Id),
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
	h.Sim = simulator.New(h.Sim.Store.CrdtType)

	helper.WriteOkNoContent(w)
}

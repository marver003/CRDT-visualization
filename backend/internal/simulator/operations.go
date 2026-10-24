package simulator

import (
	"errors"

	"github.com/google/uuid"
	"github.com/marver003/crdt/internal/types"
)

type RequestKind string

const (
	WriteRequest RequestKind = "write"
	ReadRequest  RequestKind = "read"
)

const sendDelay types.StepId = 1

var (
	ErrClientNotFound = errors.New("client does not exist")
	ErrClientBusy     = errors.New("client already sent a request in this tick")
)

type BaseOperation struct {
	StepId types.StepId
}

type ReceiveRequestOperation struct {
	BaseOperation
	Message *Message
}

type IncrementOperation struct {
	BaseOperation
	ReplicaId types.ReplicaId
	ClientId  types.ClientId
	RequestId string
	Seq       uint64
	Element   string
}

type ReadOperation struct {
	BaseOperation
	ReplicaId types.ReplicaId
	ClientId  types.ClientId
	RequestId string
	Seq       uint64
}

type SendResponseOperation struct {
	BaseOperation
	ReplicaId types.ReplicaId
	ClientId  types.ClientId
	RequestId string
	Seq       uint64
	Kind      MessageKind
	Value     uint64
	Elements  []string
}

type ReceiveResponseOperation struct {
	BaseOperation
	Message *Message
}

type SendGossipMessageOperation struct {
	BaseOperation
	From     types.ReplicaId
	To       types.ReplicaId
	ClientId types.ClientId
	Seq      uint64
}

type ReceiveGossipMessageOperation struct {
	BaseOperation
	Message *Message
}

type MergeOperation struct {
	BaseOperation
	Message *Message
}

type SendGossipAckOperation struct {
	BaseOperation
	From     types.ReplicaId
	To       types.ReplicaId
	ClientId types.ClientId
	Seq      uint64
}

type ReceiveGossipAckOperation struct {
	BaseOperation
	Message *Message
}

func (op BaseOperation) GetStepId() types.StepId {
	return op.StepId
}

func (s *Simulator) SendClientRequest(clientId types.ClientId, replicaId types.ReplicaId, kind RequestKind, element string) error {
	client, ok := s.Clients[clientId]
	if !ok {
		return ErrClientNotFound
	}

	if client.Locked(s.Now) {
		return ErrClientBusy
	}

	client.HasSent = true
	client.LastRequestAt = s.Now

	messageKind := WriteRequestMessage
	if kind == ReadRequest {
		messageKind = ReadRequestMessage
	}

	deliverAt := s.Now + s.Timing.RequestLatency
	msg := s.NewMessage(messageKind, ClientAt(clientId), NodeAt(replicaId), deliverAt)
	msg.RequestId = uuid.NewString()
	msg.Element = element
	msg.ClientId = clientId
	s.RequestSeq++
	msg.Seq = s.RequestSeq

	s.Queue.Enqueue(ReceiveRequestOperation{
		BaseOperation: BaseOperation{StepId: deliverAt},
		Message:       msg,
	})

	return nil
}

func (op ReceiveRequestOperation) Execute(sim *Simulator) error {
	sim.RemovePendingMessage(op.Message.Id)

	replicaId := types.ReplicaId(op.Message.To.Id)
	if !sim.Store.IsUp(replicaId) {
		return nil
	}

	clientId := types.ClientId(op.Message.From.Id)
	after := sim.After(sim.Timing.Processing)

	if op.Message.Kind == ReadRequestMessage {
		sim.Queue.Enqueue(ReadOperation{
			BaseOperation: after,
			ReplicaId:     replicaId,
			ClientId:      clientId,
			RequestId:     op.Message.RequestId,
			Seq:           op.Message.Seq,
		})
		return nil
	}

	sim.Queue.Enqueue(IncrementOperation{
		BaseOperation: after,
		ReplicaId:     replicaId,
		ClientId:      clientId,
		RequestId:     op.Message.RequestId,
		Seq:           op.Message.Seq,
		Element:       op.Message.Element,
	})

	return nil
}

func (op IncrementOperation) Execute(sim *Simulator) error {
	if !sim.Store.IsUp(op.ReplicaId) {
		return nil
	}

	if err := sim.Store.WriteNode(op.ReplicaId, op.Element); err != nil {
		return err
	}

	if op.ClientId != "" {
		sim.Queue.Enqueue(SendResponseOperation{
			BaseOperation: sim.After(sendDelay),
			ReplicaId:     op.ReplicaId,
			ClientId:      op.ClientId,
			RequestId:     op.RequestId,
			Seq:           op.Seq,
			Kind:          WriteAckMessage,
		})
	}

	for id := range sim.Store.Nodes {
		if id == op.ReplicaId {
			continue
		}

		sim.Queue.Enqueue(SendGossipMessageOperation{
			BaseOperation: sim.After(sendDelay),
			From:          op.ReplicaId,
			To:            id,
			ClientId:      op.ClientId,
			Seq:           op.Seq,
		})
	}

	return nil
}

func (op ReadOperation) Execute(sim *Simulator) error {
	if !sim.Store.IsUp(op.ReplicaId) {
		return nil
	}

	var value uint64
	var elements []string
	var err error

	if sim.Store.CrdtType == types.GSetType {
		elements, err = sim.Store.ReadNodeElements(op.ReplicaId)
	} else {
		value, err = sim.Store.ReadNodeValue(op.ReplicaId)
	}

	if err != nil {
		return err
	}

	sim.Queue.Enqueue(SendResponseOperation{
		BaseOperation: sim.After(sendDelay),
		ReplicaId:     op.ReplicaId,
		ClientId:      op.ClientId,
		RequestId:     op.RequestId,
		Seq:           op.Seq,
		Kind:          ReadResponseMessage,
		Value:         value,
		Elements:      elements,
	})

	return nil
}

func (op SendResponseOperation) Execute(sim *Simulator) error {
	if !sim.Store.IsUp(op.ReplicaId) {
		return nil
	}

	deliverAt := sim.Now + sim.Timing.ResponseLatency
	msg := sim.NewMessage(op.Kind, NodeAt(op.ReplicaId), ClientAt(op.ClientId), deliverAt)
	msg.RequestId = op.RequestId
	msg.ClientId = op.ClientId
	msg.Seq = op.Seq
	msg.Value = op.Value
	msg.Elements = op.Elements

	sim.Queue.Enqueue(ReceiveResponseOperation{
		BaseOperation: BaseOperation{StepId: deliverAt},
		Message:       msg,
	})

	return nil
}

func (op ReceiveResponseOperation) Execute(sim *Simulator) error {
	sim.RemovePendingMessage(op.Message.Id)

	client, ok := sim.Clients[types.ClientId(op.Message.To.Id)]
	if !ok {
		return nil
	}

	client.LastResult = &ClientResult{
		Kind:      op.Message.Kind,
		ReplicaId: types.ReplicaId(op.Message.From.Id),
		Value:     op.Message.Value,
		Elements:  op.Message.Elements,
		At:        sim.Now,
	}

	return nil
}

func (op SendGossipMessageOperation) Execute(sim *Simulator) error {
	if !sim.Store.IsUp(op.From) {
		return nil
	}

	from := sim.Store.Nodes[op.From]

	deliverAt := sim.Now + sim.Timing.GossipLatency
	msg := sim.NewMessage(GossipMessage, NodeAt(op.From), NodeAt(op.To), deliverAt)
	msg.ClientId = op.ClientId
	msg.Seq = op.Seq
	msg.State = from.Crdt.GetSnapshot()

	sim.Queue.Enqueue(ReceiveGossipMessageOperation{
		BaseOperation: BaseOperation{StepId: deliverAt},
		Message:       msg,
	})

	return nil
}

func (op ReceiveGossipMessageOperation) Execute(sim *Simulator) error {
	sim.RemovePendingMessage(op.Message.Id)

	if !sim.Store.IsUp(types.ReplicaId(op.Message.To.Id)) {
		return nil
	}

	sim.Queue.Enqueue(MergeOperation{
		BaseOperation: sim.After(sim.Timing.Merge),
		Message:       op.Message,
	})

	return nil
}

func (op MergeOperation) Execute(sim *Simulator) error {
	if !sim.Store.IsUp(types.ReplicaId(op.Message.To.Id)) {
		return nil
	}

	node := sim.Store.Nodes[types.ReplicaId(op.Message.To.Id)]

	node.Crdt.Merge(op.Message.State)

	sim.Queue.Enqueue(SendGossipAckOperation{
		BaseOperation: sim.After(sendDelay),
		From:          types.ReplicaId(op.Message.To.Id),
		To:            types.ReplicaId(op.Message.From.Id),
		ClientId:      op.Message.ClientId,
		Seq:           op.Message.Seq,
	})

	return nil
}

func (op SendGossipAckOperation) Execute(sim *Simulator) error {
	if !sim.Store.IsUp(op.From) {
		return nil
	}

	deliverAt := sim.Now + sim.Timing.GossipLatency
	msg := sim.NewMessage(GossipAckMessage, NodeAt(op.From), NodeAt(op.To), deliverAt)
	msg.ClientId = op.ClientId
	msg.Seq = op.Seq

	sim.Queue.Enqueue(ReceiveGossipAckOperation{
		BaseOperation: BaseOperation{StepId: deliverAt},
		Message:       msg,
	})

	return nil
}

func (op ReceiveGossipAckOperation) Execute(sim *Simulator) error {
	sim.RemovePendingMessage(op.Message.Id)

	return nil
}

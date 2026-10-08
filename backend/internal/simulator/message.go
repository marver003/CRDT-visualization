package simulator

import (
	"github.com/google/uuid"
	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/types"
)

type MessageKind string

const (
	GossipMessage       MessageKind = "gossip"
	WriteRequestMessage MessageKind = "write_request"
	WriteAckMessage     MessageKind = "write_ack"
	ReadRequestMessage  MessageKind = "read_request"
	ReadResponseMessage MessageKind = "read_response"
)

type EndpointKind string

const (
	NodeEndpoint   EndpointKind = "node"
	ClientEndpoint EndpointKind = "client"
)

type Endpoint struct {
	Kind EndpointKind
	Id   string
}

func NodeAt(id types.ReplicaId) Endpoint {
	return Endpoint{Kind: NodeEndpoint, Id: string(id)}
}

func ClientAt(id types.ClientId) Endpoint {
	return Endpoint{Kind: ClientEndpoint, Id: string(id)}
}

type Message struct {
	Id        string
	Kind      MessageKind
	From      Endpoint
	To        Endpoint
	RequestId string
	State     crdt.Snapshot
	Value     uint64
	SentAt    types.StepId
	DeliverAt types.StepId
}

func (s *Simulator) NewMessage(kind MessageKind, from, to Endpoint, deliverAt types.StepId) *Message {
	msg := &Message{
		Id:        uuid.NewString(),
		Kind:      kind,
		From:      from,
		To:        to,
		SentAt:    s.Now,
		DeliverAt: deliverAt,
	}

	s.PendingMessages[msg.Id] = msg

	return msg
}

func (s *Simulator) RemovePendingMessage(id string) {
	delete(s.PendingMessages, id)
}

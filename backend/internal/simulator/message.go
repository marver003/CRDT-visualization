package simulator

import (
	"github.com/google/uuid"
	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/types"
)

type Message struct {
	Id    string
	From  types.ReplicaId
	To    types.ReplicaId
	State *crdt.GCounterSnapshot
}

func (s *Simulator) NewMessage(from, to types.ReplicaId, state *crdt.GCounterSnapshot) string {
	messageId := uuid.NewString()

	s.PendingMessages[messageId] = &Message{
		Id:    messageId,
		From:  from,
		To:    to,
		State: state,
	}

	return messageId

}

func (s *Simulator) EditMessage(id string, state *crdt.GCounterSnapshot) {
	msg, ok := s.PendingMessages[id]
	if !ok {
		return
	}
	msg.State = state
	s.PendingMessages[id] = msg
}

func (s *Simulator) RemovePendingMessage(id string) {
	delete(s.PendingMessages, id)
}

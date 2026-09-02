package simulator

import (
	"github.com/google/uuid"
	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/types"
)

type Message struct {
	ID    string
	From  types.ReplicaID
	To    types.ReplicaID
	State *crdt.GCounterSnapshot
}

func (s *Simulator) NewMessage(from, to types.ReplicaID, state *crdt.GCounterSnapshot) string {
	messageId := uuid.NewString()

	s.PendingMessages[messageId] = &Message{
		ID:    messageId,
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

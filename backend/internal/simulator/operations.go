package simulator

import "github.com/marver003/crdt/internal/types"

type CreateNodeOperation struct {
	StepID    types.StepID
	ReplicaID types.ReplicaID
}

type RemoveNodeOperation struct {
	StepID    types.StepID
	ReplicaID types.ReplicaID
}

type IncrementOperation struct {
	StepID    types.StepID
	ReplicaID types.ReplicaID
}

type SendGossipMessageOperation struct {
	StepID  types.StepID
	From    types.ReplicaID
	To      types.ReplicaID
	Message *Message
}

type ReceiveGossipMessageOperation struct {
	StepID  types.StepID
	Message *Message
}

func (op CreateNodeOperation) Execute(sim *Simulator) error {
	return sim.Store.CreateNode(op.ReplicaID)
}

func (op RemoveNodeOperation) Execute(sim *Simulator) error {
	return sim.Store.RemoveNode(op.ReplicaID)
}

func (op IncrementOperation) Execute(sim *Simulator) error {
	sim.Store.IncrementNodeCounter(op.ReplicaID) // returns (crdt.GCounterSnapshot, error)

	return nil
}

func (op SendGossipMessageOperation) Execute(sim *Simulator) error {

	state := sim.Store.Nodes[op.From].Counter.GetSnapshot()

	sim.EditMessage(op.Message.ID, &state)

	return nil
}

func (op ReceiveGossipMessageOperation) Execute(sim *Simulator) error {

	node := sim.Store.Nodes[op.Message.To]

	node.Counter.Merge(op.Message.State)

	sim.RemovePendingMessage(op.Message.ID)

	return nil
}

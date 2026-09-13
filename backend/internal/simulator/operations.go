package simulator

import "github.com/marver003/crdt/internal/types"

type BaseOperation struct {
	StepId types.StepId
}

type CreateNodeOperation struct {
	BaseOperation
	ReplicaId types.ReplicaId
}

type RemoveNodeOperation struct {
	BaseOperation
	ReplicaId types.ReplicaId
}

type IncrementOperation struct {
	BaseOperation
	ReplicaId types.ReplicaId
}

type SendGossipMessageOperation struct {
	BaseOperation
	From    types.ReplicaId
	To      types.ReplicaId
	Message *Message
}

type ReceiveGossipMessageOperation struct {
	BaseOperation
	Message *Message
}

func (op BaseOperation) GetStepId() types.StepId {
	return op.StepId
}

func (op CreateNodeOperation) Execute(sim *Simulator) error {
	return sim.Store.CreateNode(op.ReplicaId)
}

func (op RemoveNodeOperation) Execute(sim *Simulator) error {
	return sim.Store.RemoveNode(op.ReplicaId)
}

func (op IncrementOperation) Execute(sim *Simulator) error {
	sim.Store.IncrementNodeCounter(op.ReplicaId) // returns (crdt.GCounterSnapshot, error)

	return nil
}

func (op SendGossipMessageOperation) Execute(sim *Simulator) error {

	state := sim.Store.Nodes[op.From].Counter.GetSnapshot()

	sim.EditMessage(op.Message.Id, &state)

	return nil
}

func (op ReceiveGossipMessageOperation) Execute(sim *Simulator) error {

	node := sim.Store.Nodes[op.Message.To]

	node.Counter.Merge(op.Message.State)

	sim.RemovePendingMessage(op.Message.Id)

	return nil
}

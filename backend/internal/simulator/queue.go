package simulator

import "github.com/marver003/crdt/internal/types"

type StepQueue struct {
	Operations map[types.StepId][]Operation
}

func NewQueue() *StepQueue {
	return &StepQueue{Operations: make(map[types.StepId][]Operation)}
}

func (q *StepQueue) Enqueue(op Operation) {
	stepId := op.GetStepId()
	q.Operations[stepId] = append(q.Operations[stepId], op)
}

func (q *StepQueue) Dequeue(stepId types.StepId) []Operation {
	operations := q.Operations[stepId]

	delete(q.Operations, stepId)

	return operations
}

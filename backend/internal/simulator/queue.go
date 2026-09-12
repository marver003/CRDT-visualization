package simulator

import "github.com/marver003/crdt/internal/types"

type StepQueue struct {
	Operations map[types.StepID][]Operation
}

func NewQueue() *StepQueue {
	return &StepQueue{Operations: make(map[types.StepID][]Operation)}
}

func (q *StepQueue) Enqueue(op Operation, stepId types.StepID) {
	q.Operations[stepId] = append(q.Operations[stepId], op)
}

func (q *StepQueue) Dequeue(stepId types.StepID) []Operation {
	operations := q.Operations[stepId]

	return operations
}

func (q *StepQueue) Size() int {
	return len(q.Operations)
}

package simulator

type StepQueue struct {
	Operations []Operation
}

func NewQueue() *StepQueue {
	return &StepQueue{Operations: []Operation{}}
}

func (q *StepQueue) Enqueue(op Operation) {
	q.Operations = append(q.Operations, op)
}

func (q *StepQueue) Dequeue() Operation {
	operation := q.Operations[0]

	q.Operations = q.Operations[1:]

	return operation
}

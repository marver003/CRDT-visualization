package simulator

import "github.com/marver003/crdt/internal/types"

type Simulator struct {
	Store           *Store
	PendingMessages map[string]*Message
	Queue           *StepQueue
	nextStepID      types.StepID
}

func New() *Simulator {
	return &Simulator{
		Store:           NewStore(),
		PendingMessages: make(map[string]*Message),
		Queue:           NewQueue(),
		nextStepID:      1,
	}
}

func (s *Simulator) GetNextStepId() types.StepID {
	stepID := s.nextStepID
	s.nextStepID++

	return stepID
}

func (s *Simulator) Step() {
	op := s.Queue.Dequeue()

	op.Execute(s)
}

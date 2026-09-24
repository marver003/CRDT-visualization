package simulator

import (
	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/types"
)

type Simulator struct {
	Store           *Store
	PendingMessages map[string]*Message
	Queue           *StepQueue
	NextStepId      types.StepId
	ExecutingStepId types.StepId
}

func New(crdtType types.CrdtType) *Simulator {
	return &Simulator{
		Store:           NewStore(crdtType),
		PendingMessages: make(map[string]*Message),
		Queue:           NewQueue(),
		NextStepId:      1,
		ExecutingStepId: 1,
	}
}

// GetNextStepId returns step ID value before it was incremented
func (s *Simulator) GetNextStepId() types.StepId {
	stepId := s.NextStepId
	s.NextStepId++

	return stepId
}

func (s *Simulator) Step() {
	ops := s.Queue.Dequeue(s.ExecutingStepId)

	for _, op := range ops {
		op.Execute(s)
	}

	s.ExecutingStepId++
}

func Load(crdtType types.CrdtType, snapshots map[types.ReplicaId]crdt.Snapshot) (*Simulator, error) {
	s := New(crdtType)

	for replicaId, snapshot := range snapshots {
		if err := s.Store.CreateNodeWithSnapshot(replicaId, snapshot); err != nil {
			return nil, err
		}
	}

	return s, nil
}

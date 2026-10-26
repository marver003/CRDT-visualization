package simulator

import (
	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/types"
)

type Timing struct {
	RequestLatency  types.StepId
	ResponseLatency types.StepId
	Processing      types.StepId
	GossipLatency   types.StepId
	Merge           types.StepId
}

func DefaultTiming() Timing {
	return Timing{
		RequestLatency:  1,
		ResponseLatency: 1,
		Processing:      1,
		GossipLatency:   1,
		Merge:           1,
	}
}

type Simulator struct {
	Store           *Store
	Clients         map[types.ClientId]*Client
	PendingMessages map[string]*Message
	PendingOps      map[PendingOpKey]*PendingOp
	Queue           *StepQueue
	Timing          Timing
	Now             types.StepId
	RequestSeq      uint64
}

func New(crdtType types.CrdtType) *Simulator {
	return &Simulator{
		Store:           NewStore(crdtType),
		Clients:         newClients(),
		PendingMessages: make(map[string]*Message),
		PendingOps:      make(map[PendingOpKey]*PendingOp),
		Queue:           NewQueue(),
		Timing:          DefaultTiming(),
	}
}

func (s *Simulator) After(delay types.StepId) BaseOperation {
	return BaseOperation{StepId: s.Now + delay}
}

func (s *Simulator) Step() {
	s.Now++

	for _, op := range s.Queue.Dequeue(s.Now) {
		op.Execute(s)
	}
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

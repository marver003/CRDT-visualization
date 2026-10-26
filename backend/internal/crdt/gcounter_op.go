package crdt

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/marver003/crdt/internal/types"
)

type CounterOp struct {
	Origin types.ReplicaId
	Seq    uint64
	Amount uint64
}

type GCounterOp struct {
	mu sync.RWMutex

	id      types.ReplicaId
	value   uint64
	nextSeq uint64
	log     []CounterOp
	applied map[types.ReplicaId]map[uint64]bool
}

type GCounterOpSnapshot struct {
	Value uint64
}

func NewGCounterOp(id types.ReplicaId) *GCounterOp {
	return &GCounterOp{
		id:      id,
		applied: make(map[types.ReplicaId]map[uint64]bool),
	}
}

func (g *GCounterOp) Type() types.CrdtType {
	return types.GCounterOpType
}

func (g *GCounterOpSnapshot) Type() types.CrdtType {
	return types.GCounterOpType
}

func (g *GCounterOp) Generate(amount uint64) CounterOp {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.nextSeq++
	op := CounterOp{Origin: g.id, Seq: g.nextSeq, Amount: amount}
	g.apply(op)
	g.log = append(g.log, op)

	return op
}

func (g *GCounterOp) Apply(op CounterOp) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.apply(op)
}

func (g *GCounterOp) apply(op CounterOp) bool {
	seen := g.applied[op.Origin]
	if seen == nil {
		seen = make(map[uint64]bool)
		g.applied[op.Origin] = seen
	}

	if seen[op.Seq] {
		return false
	}

	seen[op.Seq] = true
	g.value += op.Amount

	return true
}

func (g *GCounterOp) Have() map[types.ReplicaId]uint64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	have := make(map[types.ReplicaId]uint64, len(g.applied))

	for origin, seen := range g.applied {
		var prefix uint64
		for seen[prefix+1] {
			prefix++
		}

		have[origin] = prefix
	}

	return have
}

func (g *GCounterOp) OpsAfter(seq uint64) []CounterOp {
	g.mu.RLock()
	defer g.mu.RUnlock()

	ops := make([]CounterOp, 0)

	for _, op := range g.log {
		if op.Seq > seq {
			ops = append(ops, op)
		}
	}

	return ops
}

func (g *GCounterOp) Value() uint64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.value
}

func (g *GCounterOp) Merge(other Snapshot) {}

func (g *GCounterOp) GetSnapshot() Snapshot {
	return &GCounterOpSnapshot{Value: g.Value()}
}

func (g *GCounterOp) ApplySnapshot(state Snapshot) error {
	snapshot, ok := state.(*GCounterOpSnapshot)
	if !ok {
		return fmt.Errorf("gcounter-op: cannot apply snapshot of type %T", state)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.value = snapshot.Value

	return nil
}

func (g *GCounterOp) Marshal() ([]byte, error) {
	return json.Marshal(g.GetSnapshot())
}

func (g *GCounterOp) Unmarshal(data []byte) (Snapshot, error) {
	var snapshot GCounterOpSnapshot
	err := json.Unmarshal(data, &snapshot)

	return &snapshot, err
}

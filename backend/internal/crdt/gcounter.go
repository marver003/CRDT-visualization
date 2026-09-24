package crdt

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/marver003/crdt/internal/types"
)

type GCounter struct {
	mu sync.RWMutex

	id     types.ReplicaId
	counts map[types.ReplicaId]uint64
}

type GCounterSnapshot struct {
	Counts map[types.ReplicaId]uint64
	Value  uint64
}

func NewGCounter(id types.ReplicaId) *GCounter {
	return &GCounter{
		id:     id,
		counts: make(map[types.ReplicaId]uint64),
	}
}

func (g *GCounter) Type() types.CrdtType {
	return types.GCounterType
}

func (g *GCounterSnapshot) Type() types.CrdtType {
	return types.GCounterType
}

func (g *GCounter) Increment() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.counts[g.id]++
}

func (g *GCounter) Add(n uint64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.counts[g.id] += n
}

func (g *GCounter) Value() uint64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var sum uint64
	for _, v := range g.counts {
		sum += v
	}
	return sum
}

func (g *GCounter) Merge(other Snapshot) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for id, v := range other.(*GCounterSnapshot).Counts {
		g.counts[id] = max(v, g.counts[id])
	}
}

func (g *GCounter) GetSnapshotCounts() GCounterSnapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return GCounterSnapshot{Counts: g.counts}
}

func (g *GCounter) GetSnapshot() Snapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()

	cp := make(map[types.ReplicaId]uint64, len(g.counts))

	var total uint64 = 0
	for k, v := range g.counts {
		cp[k] = v
		total += v
	}

	return &GCounterSnapshot{Counts: cp, Value: total}
}

func (g *GCounter) ApplySnapshot(state Snapshot) error {
	snapshot, ok := state.(*GCounterSnapshot)
	if !ok {
		return fmt.Errorf("gcounter: cannot apply snapshot of type %T", state)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.counts = snapshot.Counts

	if g.counts == nil {
		g.counts = make(map[types.ReplicaId]uint64)
	}

	return nil
}

func (g *GCounter) Marshal() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return json.Marshal(g.GetSnapshotCounts())
}

func (g *GCounter) Unmarshal(data []byte) (Snapshot, error) {
	var snapshot GCounterSnapshot
	err := json.Unmarshal(data, &snapshot)

	return &snapshot, err

}

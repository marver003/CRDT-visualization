package crdt

import (
	"encoding/json"
	"sync"

	"github.com/marver003/crdt/internal/types"
)

type GCounter struct {
	mu sync.RWMutex

	id     types.ReplicaID
	counts map[types.ReplicaID]uint64
}

type GCounterSnapshot struct {
	Counts map[types.ReplicaID]uint64
	Value  uint64
}

func NewGCounter(id types.ReplicaID) *GCounter {
	return &GCounter{
		id:     id,
		counts: make(map[types.ReplicaID]uint64),
	}
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

func (g *GCounter) Merge(other *GCounterSnapshot) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for id, v := range other.Counts {
		g.counts[id] = max(v, g.counts[id])
	}
}

func (g *GCounter) GetSnapshotCounts() GCounterSnapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return GCounterSnapshot{Counts: g.counts}
}

func (g *GCounter) GetSnapshot() GCounterSnapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()

	cp := make(map[types.ReplicaID]uint64, len(g.counts))

	var total uint64 = 0
	for k, v := range g.counts {
		cp[k] = v
		total += v
	}

	return GCounterSnapshot{Counts: cp, Value: total}
}

func (g *GCounter) ApplySnapshot(state GCounterSnapshot) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.counts == nil {
		g.counts = make(map[types.ReplicaID]uint64)
	}

	for id, v := range state.Counts {
		if v > g.counts[id] {
			g.counts[id] = v
		}
	}

	return nil
}

func (g *GCounter) Marshal() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return json.Marshal(g.GetSnapshotCounts())
}

func (g *GCounter) Unmarshal(data []byte) (GCounterSnapshot, error) {
	var snapshot GCounterSnapshot
	err := json.Unmarshal(data, &snapshot)

	return snapshot, err

}

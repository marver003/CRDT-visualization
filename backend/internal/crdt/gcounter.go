package crdt

import (
	"maps"

	"github.com/marver003/crdt/internal/types"
)

type GCounter struct {
	id     types.ReplicaID
	counts map[types.ReplicaID]uint64
}

type GCounterState struct {
	Counts map[types.ReplicaID]uint64
}

func NewGCounter(id types.ReplicaID) *GCounter {
	return &GCounter{
		id:     id,
		counts: make(map[types.ReplicaID]uint64),
	}
}

func (g *GCounter) Increment() {
	g.counts[g.id]++
}

func (g *GCounter) Add(n uint64) {
	g.counts[g.id] += n
}

func (g *GCounter) Value() uint64 {
	var sum uint64
	for _, v := range g.counts {
		sum += v
	}
	return sum
}

func (g *GCounter) Merge(other *GCounter) {
	for id, v := range other.counts {
		if v > g.counts[id] {
			g.counts[id] = v
		}
	}
}

func (g *GCounter) Clone() *GCounter {
	cp := &GCounter{
		id:     g.id,
		counts: make(map[types.ReplicaID]uint64, len(g.counts)),
	}

	for id, v := range g.counts {
		cp.counts[id] = v
	}

	return cp
}

func (g *GCounter) State() GCounterState {
	cp := make(map[types.ReplicaID]uint64, len(g.counts))
	maps.Copy(cp, g.counts)
	return GCounterState{Counts: cp}
}

func (g *GCounter) ApplyState(state GCounterState) error {
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

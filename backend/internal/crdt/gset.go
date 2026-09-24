package crdt

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/marver003/crdt/internal/types"
)

type GSet struct {
	mu sync.RWMutex

	id  types.ReplicaId
	set types.Set[string]
}

type GSetSnapshot struct {
	Set types.Set[string]
}

func NewGSet(id types.ReplicaId) *GSet {
	return &GSet{
		id:  id,
		set: make(types.Set[string]),
	}
}

func (g *GSet) Type() types.CrdtType {
	return types.GSetType
}

func (g *GSetSnapshot) Type() types.CrdtType {
	return types.GSetType
}

func (g *GSet) Add(n string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.set[n] = struct{}{}
}

func (g *GSet) Merge(otherSnapshot Snapshot) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for element := range otherSnapshot.(*GSetSnapshot).Set {
		g.set[element] = struct{}{}
	}
}

func (g *GSet) GetSnapshot() Snapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()

	cp := make(types.Set[string], len(g.set))

	for element := range g.set {
		cp[element] = struct{}{}
	}

	return &GSetSnapshot{Set: cp}
}

func (g *GSet) ApplySnapshot(snapshot Snapshot) error {
	gsetSnapshot, ok := snapshot.(*GSetSnapshot)
	if !ok {
		return fmt.Errorf("gset: cannot apply snapshot of type %T", snapshot)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.set = gsetSnapshot.Set

	if g.set == nil {
		g.set = make(types.Set[string])
	}

	return nil
}

func (g *GSet) Marshal() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return json.Marshal(g.GetSnapshot())
}

func (g *GSet) Unmarshal(data []byte) (Snapshot, error) {
	var snapshot GSetSnapshot
	err := json.Unmarshal(data, &snapshot)

	return &snapshot, err

}

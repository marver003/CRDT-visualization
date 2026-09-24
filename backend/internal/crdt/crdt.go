package crdt

import (
	"fmt"

	"github.com/marver003/crdt/internal/types"
)

type CRDT interface {
	Type() types.CrdtType

	Merge(snapshot Snapshot)
	GetSnapshot() Snapshot
	ApplySnapshot(snapshot Snapshot) error

	Marshal() ([]byte, error)
	Unmarshal(snapshotBytes []byte) (Snapshot, error)
}

type Snapshot interface {
	Type() types.CrdtType
}

// NewCrdt creates an empty CRDT of the given type for replica id.
func NewCrdt(id types.ReplicaId, crdtType types.CrdtType) (CRDT, error) {
	switch crdtType {
	case types.GCounterType:
		return NewGCounter(id), nil
	case types.GSetType:
		return NewGSet(id), nil
	default:
		return nil, fmt.Errorf("unsupported crdt type %q", crdtType)
	}
}

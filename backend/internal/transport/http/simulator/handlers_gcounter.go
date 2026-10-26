package simhttp

import (
	"fmt"

	"github.com/marver003/crdt/internal/simulator"
	"github.com/marver003/crdt/internal/types"
)

const maxGCounterAmount uint64 = 1000

func gcounterAmount(crdtType types.CrdtType, kind simulator.RequestKind, amount uint64) (uint64, error) {
	if (crdtType != types.GCounterType && crdtType != types.GCounterOpType) || kind != simulator.WriteRequest {
		return 0, nil
	}

	if amount == 0 {
		return 1, nil
	}

	if amount > maxGCounterAmount {
		return 0, fmt.Errorf(`"amount" must be between 1 and %d`, maxGCounterAmount)
	}

	return amount, nil
}

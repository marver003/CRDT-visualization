package simhttp

import (
	"errors"
	"strings"

	"github.com/marver003/crdt/internal/simulator"
	"github.com/marver003/crdt/internal/types"
)

func stateFor(store *simulator.Store) (any, error) {
	if store.CrdtType == types.GSetType {
		return store.GetGSetState()
	}

	return store.GetState()
}

func writeStepType(crdtType types.CrdtType) string {
	if crdtType == types.GSetType {
		return "add"
	}

	return "increment"
}

func validateGSetWrite(crdtType types.CrdtType, kind simulator.RequestKind, element string) error {
	if crdtType != types.GSetType || kind != simulator.WriteRequest {
		return nil
	}

	if strings.TrimSpace(element) == "" {
		return errors.New(`Request body must be JSON with a non-empty "element" field for a gset write`)
	}

	return nil
}

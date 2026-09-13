package simulator

import "github.com/marver003/crdt/internal/types"

type Operation interface {
	GetStepId() types.StepId
	Execute(*Simulator) error
}

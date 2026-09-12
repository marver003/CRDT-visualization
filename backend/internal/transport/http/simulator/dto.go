package simhttp

import "github.com/marver003/crdt/internal/types"

type reqCreateNode struct {
	ID types.ReplicaID `json:"id"`
}

type reqIncrementNode struct {
	ID types.ReplicaID `json:"id"`
}

type reqMergeNodes struct {
	SourceId types.ReplicaID `json:"source-id"`
	TargetId types.ReplicaID `json:"target-id"`
}

type respGetState struct {
	State map[types.ReplicaID]map[types.ReplicaID]uint64 `json:"state"`
}

type respGetSteps struct {
	Steps []respStep `json:"steps"`
}

type respStep struct {
	StepID types.StepID `json:"stepId"`
	Type   string       `json:"type"`

	ReplicaID types.ReplicaID `json:"replicaId,omitempty"`

	From types.ReplicaID `json:"from,omitempty"`
	To   types.ReplicaID `json:"to,omitempty"`
}

type reqLoad respGetState

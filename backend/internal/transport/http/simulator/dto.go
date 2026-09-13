package simhttp

import "github.com/marver003/crdt/internal/types"

type reqCreateNode struct {
	Id types.ReplicaId `json:"id"`
}

type reqIncrementNode struct {
	Id types.ReplicaId `json:"id"`
}

type reqMergeNodes struct {
	SourceId types.ReplicaId `json:"source-id"`
	TargetId types.ReplicaId `json:"target-id"`
}

type respGetState struct {
	State map[types.ReplicaId]map[types.ReplicaId]uint64 `json:"state"`
}

type respGetSteps struct {
	Steps []respStep `json:"steps"`
}

type respStep struct {
	StepId types.StepId `json:"stepId"`
	Type   string       `json:"type"`

	ReplicaId types.ReplicaId `json:"replicaId,omitempty"`

	From types.ReplicaId `json:"from,omitempty"`
	To   types.ReplicaId `json:"to,omitempty"`
}

type reqLoad respGetState

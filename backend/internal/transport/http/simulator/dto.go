package simhttp

import "github.com/marver003/crdt/internal/types"

type reqCreateNode struct {
	Id types.ReplicaId `json:"id"`
}

type reqClientRequest struct {
	ClientId types.ClientId  `json:"clientId"`
	NodeId   types.ReplicaId `json:"nodeId"`
	Element  string          `json:"element,omitempty"`
	Amount   uint64          `json:"amount,omitempty"`
}

type reqMergeNodes struct {
	SourceId types.ReplicaId `json:"source-id"`
	TargetId types.ReplicaId `json:"target-id"`
}

type reqSwitchSimulator struct {
	Crdt types.CrdtType `json:"crdt"`
}

type respGetState struct {
	Crdt     types.CrdtType    `json:"crdt"`
	Now      types.StepId      `json:"now"`
	State    any               `json:"state"`
	Disabled []types.ReplicaId `json:"disabled"`
	Clients  []respClient      `json:"clients"`
	InFlight []respMessage     `json:"inFlight"`
}

type respClient struct {
	Id     types.ClientId    `json:"id"`
	Locked bool              `json:"locked"`
	Result *respClientResult `json:"result"`
}

type respClientResult struct {
	Kind     string          `json:"kind"`
	NodeId   types.ReplicaId `json:"nodeId"`
	Value    uint64          `json:"value"`
	Elements []string        `json:"elements,omitempty"`
	At       types.StepId    `json:"at"`
}

type respEndpoint struct {
	Kind string `json:"kind"`
	Id   string `json:"id"`
}

type respMessage struct {
	Id        string         `json:"id"`
	Kind      string         `json:"kind"`
	From      respEndpoint   `json:"from"`
	To        respEndpoint   `json:"to"`
	SentAt    types.StepId   `json:"sentAt"`
	DeliverAt types.StepId   `json:"deliverAt"`
	ClientId  types.ClientId `json:"clientId,omitempty"`
	Seq       uint64         `json:"seq,omitempty"`
}

type respGetSteps struct {
	Steps []respStep `json:"steps"`
}

type respStep struct {
	StepId types.StepId `json:"stepId"`
	Type   string       `json:"type"`

	ReplicaId types.ReplicaId `json:"replicaId,omitempty"`
	ClientId  types.ClientId  `json:"clientId,omitempty"`
	Seq       uint64          `json:"seq,omitempty"`
	Element   string          `json:"element,omitempty"`

	From types.ReplicaId `json:"from,omitempty"`
	To   types.ReplicaId `json:"to,omitempty"`
}

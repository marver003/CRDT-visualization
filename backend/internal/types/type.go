package types

type CrdtType string

const (
	GCounterType   CrdtType = "gcounter"
	GSetType       CrdtType = "gset"
	GCounterOpType CrdtType = "gcounter-op"
	LWWSetType     CrdtType = "lwwset"
)

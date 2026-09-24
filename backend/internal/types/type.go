package types

type CrdtType string

const (
	GCounterType CrdtType = "gcounter"
	GSetType     CrdtType = "gset"
	LWWSetType   CrdtType = "lwwset"
)

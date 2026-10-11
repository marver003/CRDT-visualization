package simulator

import "github.com/marver003/crdt/internal/types"

var DefaultClientIds = []types.ClientId{"1", "2"}

type ClientResult struct {
	Kind      MessageKind
	ReplicaId types.ReplicaId
	Value     uint64
	At        types.StepId
}

type Client struct {
	Id            types.ClientId
	LastResult    *ClientResult
	HasSent       bool
	LastRequestAt types.StepId
}

func (c *Client) Locked(now types.StepId) bool {
	return c.HasSent && c.LastRequestAt == now
}

func newClients() map[types.ClientId]*Client {
	clients := make(map[types.ClientId]*Client, len(DefaultClientIds))

	for _, id := range DefaultClientIds {
		clients[id] = &Client{Id: id}
	}

	return clients
}

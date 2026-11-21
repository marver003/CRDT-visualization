package live

import (
	"log"

	"github.com/marver003/crdt/internal/node"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	Node *node.Node
}

// NewHandler returns a Handler backed by the given Node.
func NewHandler(n *node.Node) *Handler {
	log.Println("Creating new handler")
	return &Handler{Node: n}
}

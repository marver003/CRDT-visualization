package live

import (
	"fmt"
	"log"
	"net/http"

	"github.com/marver003/crdt/internal/types"
)

// corsMiddleware adds permissive CORS headers so a browser-based frontend
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*") // for development, for production actual frontend domain should be set
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle CORS pre-flight requests.
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func registerRoutes(mux *http.ServeMux, h *Handler) error {
	switch crdtType := h.Node.Crdt.Type(); crdtType {
	case types.GCounterType:
		mux.HandleFunc("GET /replicaState", h.GetCounterState)
		mux.HandleFunc("POST /increment", h.IncrementCounter)
	case types.GSetType:
		mux.HandleFunc("GET /replicaState", h.GetSetState)
		mux.HandleFunc("POST /add", h.AddElement)
	default:
		return fmt.Errorf("crdt type %q is not supported by the live node (supported: %s, %s)", crdtType, types.GCounterType, types.GSetType)
	}

	return nil
}

// NewRouter builds the HTTP mux for the node's CRDT type, or fails if the live node does not support that type.
func NewRouter(h *Handler) (http.Handler, error) {
	log.Println("Creating new router")
	mux := http.NewServeMux()

	if err := registerRoutes(mux, h); err != nil {
		return nil, err
	}

	return loggingMiddleware(corsMiddleware(mux)), nil
}

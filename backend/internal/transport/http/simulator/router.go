package simhttp

import (
	"log"
	"net/http"
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

func registerRoutes(mux *http.ServeMux, h *Handler) {

	mux.HandleFunc("POST /createNode", h.CreateNode)               // body should contain replicaId
	mux.HandleFunc("DELETE /removeNode/{replicaId}", h.RemoveNode) // no body
	mux.HandleFunc("GET /state", h.GetState)                       // state for everything
	mux.HandleFunc("POST /increment", h.IncrementNode)             // body should contain replicaId
	mux.HandleFunc("POST /merge", h.MergeNodes)                    // body should  contain sourceReplicaId and targetReplicaId
	mux.HandleFunc("GET /steps", h.GetSteps)                       // no body
	mux.HandleFunc("POST /step", h.NextStep)                       // no body
	mux.HandleFunc("POST /reset", h.Reset)                         // no body

}

// NewRouter builds and returns the HTTP mux for the simulator API.
func NewRouter(h *Handler) http.Handler {
	log.Println("Creating new SIMULATOR router")
	mux := http.NewServeMux()

	registerRoutes(mux, h)

	return loggingMiddleware(corsMiddleware(mux))
}

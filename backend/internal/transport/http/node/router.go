package internal

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

	mux.HandleFunc("GET /replicaState", h.GetReplicaState)
	mux.HandleFunc("POST /increment", h.IncrementReplica)
}

// NewRouter builds and returns the HTTP mux for the GCounter API.
func NewRouter(h *Handler) http.Handler {
	log.Println("Creating new router")
	mux := http.NewServeMux()

	registerRoutes(mux, h)

	return loggingMiddleware(corsMiddleware(mux))
}

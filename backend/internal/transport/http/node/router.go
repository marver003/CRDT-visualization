package internal

import (
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
	mux.HandleFunc("POST /replicas", h.CreateReplica) // Create a new replica
	mux.HandleFunc("GET /replicas", h.ListReplicas)   // list all replicas

	mux.HandleFunc("GET /replicas/{id}", h.GetReplica)       // get replica state
	mux.HandleFunc("DELETE /replicas/{id}", h.DeleteReplica) // delete a replica

	mux.HandleFunc("POST /replicas/{id}/increment", h.IncrementReplica) // increment
	mux.HandleFunc("POST /replicas/{id}/add", h.AddToReplica)           // add N
	mux.HandleFunc("POST /replicas/{id}/merge", h.MergeReplica)         // merge state
}

// NewRouter builds and returns the HTTP mux for the GCounter API.
func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	registerRoutes(mux, h)

	return loggingMiddleware(corsMiddleware(mux))
}

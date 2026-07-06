package internal

import (
	"log"
	"net/http"
	"time"
)

// responseRecorder wraps http.ResponseWriter so we can capture the status code
// for loggin purposes
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

// Write satisfies http.ResponseWriter; if WriteHeader was never called explicitly
// the standard library defaults to 200
func (rr *responseRecorder) Write(b []byte) (int, error) {
	if rr.status == 0 {
		rr.status = http.StatusOK
	}
	return rr.ResponseWriter.Write(b)
}

// loggingMiddleware logs every incoming request with method, path,
// response status code, and elapsed time.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()

		respRecorder := &responseRecorder{ResponseWriter: writer, status: 0}
		next.ServeHTTP(respRecorder, request)

		status := respRecorder.status
		if status == 0 {
			status = http.StatusOK
		}

		log.Printf("%-6s %-40s %d  %s",
			request.Method,
			request.URL.RequestURI(),
			status,
			time.Since(start).Round(time.Microsecond),
		)
	})
}

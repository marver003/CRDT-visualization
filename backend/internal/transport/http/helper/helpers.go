package helper

import (
	"encoding/json"
	"net/http"
)

// WriteJSON serialises v as JSON and writes it with the given HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteOkNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WriteError writes a JSON error envelope with the given HTTP status code.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// func ParseReplicaId(id string) (types.ReplicaID, error) {
// 	intId, err := strconv.ParseUint(id, 10, 64)
// 	if err != nil {
// 		return "", err
// 	}
// 	return types.ReplicaID(intId), nil
// }

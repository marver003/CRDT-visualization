package internal

import (
	"strconv"

	"github.com/marver003/crdt/internal/types"
)

func parseReplicaId(id string) (types.ReplicaID, error) {
	intId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, err
	}
	return types.ReplicaID(intId), err
}

package cluster_id

import "github.com/google/uuid"

var (
	instanceId = uuid.New().String()
)

func GetInstanceID() string {
	return instanceId
}

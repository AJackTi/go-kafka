package events

import "github.com/AJackTi/go-kafka/pkg/es"

const (
	TaskDeletedEventType es.EventType = "TASK_DELETED_V1"
)

type TaskDeletedEventV1 struct {
	ID          string `json:"id"`
	Metadata    []byte `json:"-"`
}

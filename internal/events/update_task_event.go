package events

import "github.com/AJackTi/go-kafka/pkg/es"

const (
	TaskUpdatedEventType es.EventType = "TASK_UPDATED_V1"
)

type TaskUpdatedEventV1 struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Metadata    []byte `json:"-"`
}

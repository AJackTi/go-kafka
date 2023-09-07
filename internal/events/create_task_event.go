package events

import (
	"github.com/evrone/go-clean-template/pkg/es"
)

const (
	TaskCreatedEventType es.EventType = "TASK_CREATED_V1"
)

type TaskCreatedEventV1 struct {
	Title       string    `json:"title"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Metadata  	[]byte    `json:"-"`
}
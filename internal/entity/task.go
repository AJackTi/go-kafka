package entity

import (
	"time"
)

type Task struct {
	ID          string    `json:"id,omitempty"`
	AggregateID string    `json:"aggregateID,omitempty"`
	Title       string    `json:"title,omitempty"`
	Name        string    `json:"name,omitempty"`
	Image       string    `json:"image,omitempty"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

func NewTask(id string) *Task {
	return &Task{AggregateID: id}
}

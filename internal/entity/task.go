package entity

import (
	"time"
)

type Task struct {
	AggregateID string       `json:"aggregateID"`
	Title       string    `json:"title"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewTask(id string) *Task {
	return &Task{AggregateID: id}
}
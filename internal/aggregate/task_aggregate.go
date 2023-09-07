package aggregate

import (
	"errors"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/events"
	"github.com/evrone/go-clean-template/pkg/es"
)

const (
	TaskAggregateType es.AggregateType = "Task"
)

var (
	ErrUnknownEventType = errors.New("unknown event type")
)

// TaskAggregate
type TaskAggregate struct {
	*es.AggregateBase
	Task *entity.Task
}

func (a *TaskAggregate) When(event any) error {
	switch evt := event.(type) {
	case *events.TaskCreatedEventV1:
		a.Task.Title = evt.Title
		a.Task.Name = evt.Name
		a.Task.Image = evt.Image
		a.Task.Description = evt.Description
		a.Task.Status = evt.Status
		return nil

	default:
		return ErrUnknownEventType
	}
}

func NewTaskAggregate(id string) *TaskAggregate {
	if id == "" {
		return nil
	}

	taskAggregate := &TaskAggregate{Task: entity.NewTask(id)}
	aggregateBase := es.NewAggregateBase(taskAggregate.When)
	aggregateBase.SetType(TaskAggregateType)
	aggregateBase.SetID(id)
	taskAggregate.AggregateBase = aggregateBase
	return taskAggregate
}

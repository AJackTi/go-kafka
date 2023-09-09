package usecase

import (
	"context"

	"github.com/AJackTi/go-kafka/internal/aggregate"
	"github.com/AJackTi/go-kafka/internal/domain"
	"github.com/AJackTi/go-kafka/internal/entity"
	internalEvent "github.com/AJackTi/go-kafka/internal/events"
	"github.com/AJackTi/go-kafka/pkg/es"
	"github.com/google/uuid"
)

type TaskUseCase struct {
	eventSerializer *domain.EventSerializer
	eventBus        *es.KafkaEventsBus
}

// NewTask -.
func NewTask(eventSerializer *domain.EventSerializer, eventBus *es.KafkaEventsBus) *TaskUseCase {
	return &TaskUseCase{
		eventSerializer: eventSerializer,
		eventBus:        eventBus,
	}
}

type CreateTaskRequest struct {
	Title       string
	Name        string
	Image       string
	Description string
	Status      string
}

// CreateTask - Create task.
func (uc *TaskUseCase) CreateTask(ctx context.Context, request *CreateTaskRequest) error {
	// Push to Kafka
	taskAggregate := aggregate.NewTaskAggregate(uuid.Must(uuid.NewRandom()).String())
	taskAggregate.Task = &entity.Task{
		Title:       request.Title,
		Name:        request.Name,
		Image:       request.Image,
		Description: request.Description,
		Status:      request.Status,
	}
	taskCreatedEvent := &internalEvent.TaskCreatedEventV1{
		Title:       request.Title,
		Name:        request.Name,
		Image:       request.Image,
		Description: request.Description,
		Status:      request.Status,
	}
	event, err := uc.eventSerializer.SerializeEvent(taskAggregate, taskCreatedEvent)
	if err != nil {
		return err
	}

	if err := uc.eventBus.ProcessEvents(ctx, []es.Event{event}); err != nil {
		return err
	}

	return nil
}

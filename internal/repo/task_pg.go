package repo

import (
	"context"

	"github.com/evrone/go-clean-template/internal/aggregate"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/evrone/go-clean-template/internal/entity"
	internalEvent "github.com/evrone/go-clean-template/internal/events"
	"github.com/evrone/go-clean-template/pkg/es"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/google/uuid"
)

// TaskRepo -.
type TaskRepo struct {
	pg              *postgres.Postgres
	eventBus        *es.KafkaEventsBus
	eventSerializer *domain.EventSerializer
}

// NewTask -.
func NewTask(pg *postgres.Postgres, eventBus *es.KafkaEventsBus, eventSerializer *domain.EventSerializer) *TaskRepo {
	return &TaskRepo{
		pg:              pg,
		eventBus:        eventBus,
		eventSerializer: eventSerializer,
	}
}

// CreateTask -.
func (r *TaskRepo) CreateTask(ctx context.Context, task *entity.Task) error {
	// sql, args, err := r.pg.Builder.
	// 	Insert("tasks").
	// 	Columns("title, name, image, description, status").
	// 	Values(t.Title, t.Name, t.Image, t.Description, t.Status).
	// 	ToSql()
	// if err != nil {
	// 	return fmt.Errorf("TaskRepo - CreateTask - r.Builder: %w", err)
	// }

	// _, err = r.pg.Pool.Exec(ctx, sql, args...)
	// if err != nil {
	// 	return fmt.Errorf("TaskRepo - CreateTask - r.Pool.Exec: %w", err)
	// }

	// TODO: save into database event sourcing

	// Push to Kafka
	taskAggregate := aggregate.NewTaskAggregate(uuid.Must(uuid.NewRandom()).String())
	taskAggregate.Task = task
	taskCreatedEvent := &internalEvent.TaskCreatedEventV1{}
	event, err := r.eventSerializer.SerializeEvent(taskAggregate, taskCreatedEvent)
	if err != nil {
		return err
	}

	if err := r.eventBus.ProcessEvents(ctx, []es.Event{event}); err != nil {
		return err
	}

	return nil
}

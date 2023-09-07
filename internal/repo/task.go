package repo

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/postgres"
)

// TaskRepo -.
type TaskRepo struct {
	pg *postgres.Postgres
}

// NewTask -.
func NewTask(pg *postgres.Postgres) *TaskRepo {
	return &TaskRepo{
		pg: pg,
	}
}

// CreateTask -.
func (r *TaskRepo) CreateTask(ctx context.Context, task *entity.Task) error {
	sql, args, err := r.pg.Builder.
		Insert("tasks").
		Columns("title, name, image, description, status").
		Values(task.Title, task.Name, task.Image, task.Description, task.Status).
		ToSql()
	if err != nil {
		return fmt.Errorf("TaskRepo - CreateTask - r.Builder: %w", err)
	}

	_, err = r.pg.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("TaskRepo - CreateTask - r.Pool.Exec: %w", err)
	}

	return nil
}

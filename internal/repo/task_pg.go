package repo

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/postgres"
)

// TaskRepo -.
type TaskRepo struct {
	*postgres.Postgres
}

// NewTask -.
func NewTask(pg *postgres.Postgres) *TaskRepo {
	return &TaskRepo{pg}
}

// CreateTask -.
func (r *TaskRepo) CreateTask(ctx context.Context, t *entity.Task) error {
	sql, args, err := r.Builder.
		Insert("tasks").
		Columns("title, name, image, description, status").
		Values(t.Title, t.Name, t.Image, t.Description, t.Status).
		ToSql()
	if err != nil {
		return fmt.Errorf("TaskRepo - CreateTask - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("TaskRepo - CreateTask - r.Pool.Exec: %w", err)
	}

	return nil
}

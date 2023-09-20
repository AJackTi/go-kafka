package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AJackTi/go-kafka/internal/entity"
)

// TaskRepo -.
type TaskRepo struct {
	db *sql.DB
}

// NewTask -.
func NewTask(db *sql.DB) *TaskRepo {
	return &TaskRepo{
		db: db,
	}
}

// CreateTask -.
func (r *TaskRepo) CreateTask(ctx context.Context, task *entity.Task) error {
	sqlStatement := `
	INSERT INTO tasks (title, name, image, description, status)
	VALUES ($1, $2, $3, $4, $5)`
	result, err := r.db.Exec(sqlStatement, task.Title, task.Name, task.Image, task.Description, task.Status)
	if err != nil {
		return fmt.Errorf("TaskRepo - CreateTask - r.Exec: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("TaskRepo - CreateTask - RowsAffected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("TaskRepo - CreateTask - rowsAffected wrong")
	}

	return nil
}

// UpdateTask -.
func (r *TaskRepo) UpdateTask(ctx context.Context, task *entity.Task) error {
	sqlStatement := `
	UPDATE tasks
	SET title = ?, image = ?, name = ?, description = ?, status = ?)
	WHERE id = ?`

	result, err := r.db.Exec(sqlStatement, task.Title, task.Image, task.Name, task.Description, task.Status, task.ID)
	if err != nil {
		return fmt.Errorf("TaskRepo - UpdateTask - r.Exec: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("TaskRepo - UpdateTask - RowsAffected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("TaskRepo - UpdateTask - rowsAffected wrong")
	}

	return nil
}

// DeleteTask -.
func (r *TaskRepo) DeleteTask(ctx context.Context, task *entity.Task) error {
	sqlStatement := `
	DELETE FROM tasks WHERE id = ?`

	result, err := r.db.Exec(sqlStatement, task.ID)
	if err != nil {
		return fmt.Errorf("TaskRepo - DeleteTask - r.Exec: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("TaskRepo - DeleteTask - RowsAffected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("TaskRepo - DeleteTask - rowsAffected wrong")
	}

	return nil
}

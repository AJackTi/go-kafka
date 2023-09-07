package usecase

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity"
)

type TaskUseCase struct {
	repo TaskRepo
}

// NewTask -.
func NewTask(r TaskRepo) *TaskUseCase {
	return &TaskUseCase{
		repo: r,
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
func (uc *TaskUseCase) CreateTask(ctx context.Context, task *CreateTaskRequest) error {
	if err := uc.repo.CreateTask(ctx, &entity.Task{
		Title:       task.Title,
		Name:        task.Name,
		Image:       task.Image,
		Description: task.Description,
		Status:      task.Status,
	}); err != nil {
		return fmt.Errorf("TaskUseCase - CreateTask - s.repo.Create: %w", err)
	}

	return nil
}

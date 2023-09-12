// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/AJackTi/go-kafka/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// TaskUsecase
	TaskUsecase interface {
		CreateTask(context.Context, *CreateTaskRequest) error
		UpdateTask(context.Context, string, *UpdateTaskRequest) error
		DeleteTask(context.Context, string) error
		// List(context.Context) ([]*entity.Task, error)
		// Get(context.Context, string) (*entity.Task, error)
	}

	// TaskRepo -.
	TaskRepo interface {
		CreateTask(context.Context, *entity.Task) error
		UpdateTask(context.Context, string, *entity.Task) error
		DeleteTask(context.Context, *entity.Task) error
		// List(context.Context) ([]*entity.Task, error)
		// Get(context.Context, string) (*entity.Task, error)
	}
)

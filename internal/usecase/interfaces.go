// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// Translation -.
	Translation interface {
		Translate(context.Context, entity.Translation) (entity.Translation, error)
		History(context.Context) ([]entity.Translation, error)
	}

	// Task
	Task interface {
		CreateTask(context.Context, *CreateTaskRequest) error
		// List(context.Context) ([]*entity.Task, error)
		// Get(context.Context, string) (*entity.Task, error)
		// Update(context.Context, string, *entity.Task) error
		// Delete(context.Context, string) error
	}

	// TranslationRepo -.
	TranslationRepo interface {
		Store(context.Context, entity.Translation) error
		GetHistory(context.Context) ([]entity.Translation, error)
	}

	// TaskRepo -.
	TaskRepo interface {
		CreateTask(context.Context, *entity.Task) error
		// List(context.Context) ([]*entity.Task, error)
		// Get(context.Context, string) (*entity.Task, error)
		// Update(context.Context, string, *entity.Task) error
		// Delete(context.Context, string) error
	}

	// TranslationWebAPI -.
	TranslationWebAPI interface {
		Translate(entity.Translation) (entity.Translation, error)
	}
)

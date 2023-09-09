package v1

import "github.com/AJackTi/go-kafka/internal/usecase"

type handler struct {
	taskUc *usecase.TaskUseCase
}

func New(taskUc *usecase.TaskUseCase) *handler {
	return &handler{
		taskUc: taskUc,
	}
}

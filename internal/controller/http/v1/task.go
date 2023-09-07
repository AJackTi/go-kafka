package v1

import (
	"net/http"

	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
)

type taskRoutes struct {
	t usecase.Task
	l logger.Interface
}

func newTaskRoutes(handler *gin.RouterGroup, t usecase.Task, l logger.Interface) {
	r := &taskRoutes{t, l}

	h := handler.Group("/tasks")
	{
		h.POST("", r.CreateTask)
	}
}

type RequestCreateTask struct {
	Title       string `json:"title"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// @Summary     Create task
// @Description Create new task
// @ID          task
// @Tags  	    create_task
// @Accept      json
// @Produce     json
// @Success     200
// @Failure     500 {object} response
// @Router      /tasks [post]
func (r *taskRoutes) CreateTask(c *gin.Context) {
	var request RequestCreateTask
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err := r.t.CreateTask(c.Request.Context(), &usecase.CreateTaskRequest{
		Title:       request.Title,
		Name:        request.Name,
		Image:       request.Image,
		Description: request.Description,
		Status:      request.Status,
	})
	if err != nil {
		r.l.Error(err, "http - v1 - create_task")
		errorResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, "")
}

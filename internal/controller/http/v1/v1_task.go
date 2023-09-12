package v1

import (
	"errors"
	"net/http"

	"github.com/AJackTi/go-kafka/internal/usecase"
	"github.com/AJackTi/go-kafka/pkg/logger"
	"github.com/gin-gonic/gin"
)

type taskRoutes struct {
	taskUc usecase.TaskUsecase
	logger logger.Interface
}

func (hand *handler) NewTaskRoutes(handler *gin.RouterGroup, taskUc usecase.TaskUsecase, logger logger.Interface) *handler {
	router := &taskRoutes{taskUc, logger}

	hl := handler.Group("/tasks")
	{
		hl.POST("", router.CreateTask)
		hl.GET("", router.List)
		hl.PUT("/:id", router.UpdateTask)
		hl.DELETE("/:id", router.DeleteTask)
	}

	return hand
}

type RequestCreateTask struct {
	Title       string `json:"title"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type RequestUpdateTask struct {
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

	err := r.taskUc.CreateTask(c.Request.Context(), &usecase.CreateTaskRequest{
		Title:       request.Title,
		Name:        request.Name,
		Image:       request.Image,
		Description: request.Description,
		Status:      request.Status,
	})
	if err != nil {
		r.logger.Error(err, "http - v1 - create_task")
		errorResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, request)
}

// @Summary     List task
// @Description List all the tasks
// @ID          task
// @Tags  	    list_task
// @Accept      json
// @Produce     json
// @Success     200
// @Failure     500 {object} response
// @Router      /tasks [post]
func (r *taskRoutes) List(c *gin.Context) {
	c.JSON(http.StatusOK, "")
}

// @Summary     Update task
// @Description Update task
// @ID          task
// @Tags  	    update_task
// @Accept      json
// @Produce     json
// @Success     200
// @Failure     500 {object} response
// @Router      /tasks [post]
func (r *taskRoutes) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid request"))
		return
	}

	var request RequestUpdateTask
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err := r.taskUc.UpdateTask(c.Request.Context(), taskID, &usecase.UpdateTaskRequest{
		Title:       request.Title,
		Name:        request.Name,
		Image:       request.Image,
		Description: request.Description,
		Status:      request.Status,
	})
	if err != nil {
		r.logger.Error(err, "http - v1 - update_task")
		errorResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, request)
}

// @Summary     Delete task
// @Description Delete task
// @ID          task
// @Tags  	    delete_task
// @Accept      json
// @Produce     json
// @Success     200
// @Failure     500 {object} response
// @Router      /tasks [delete]
func (r *taskRoutes) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid request"))
		return
	}
	err := r.taskUc.DeleteTask(c.Request.Context(), taskID)
	if err != nil {
		r.logger.Error(err, "http - v1 - delete_task")
		errorResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, "ok")
}

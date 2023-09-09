// Package v1 implements routing paths. Each services in own file.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Swagger docs.
	"github.com/evrone/go-clean-template/config"
	_ "github.com/evrone/go-clean-template/docs"
	v1 "github.com/evrone/go-clean-template/internal/controller/http/v1"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/es"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/postgres"
)

// NewRouter -.
// Swagger spec:
// @title       Go Clean Template API
// @description Using a service as an example
// @version     1.0
// @host        localhost:8080
// @BasePath    /v1
func NewRouter(cfg *config.Config,
	handler *gin.Engine,
	log logger.Interface,
	pg *postgres.Postgres,
	eventSerializer *domain.EventSerializer,
	eventBus *es.KafkaEventsBus) {
	// Options
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())

	// Swagger
	if cfg.App.Env != "production" {
		swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
		handler.GET("/swagger/*any", swaggerHandler)
	}

	// K8s probe
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Prometheus metrics
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Routers
	handlerGroup := handler.Group("/api/v1")
	{
		// Use case
		taskUc := usecase.NewTask(eventSerializer, eventBus)

		// handler
		handlerController := v1.New(taskUc)

		handlerController.NewTaskRoutes(handlerGroup, taskUc, log)
	}
}

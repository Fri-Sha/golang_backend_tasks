package handlers

import (
	"main/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{repository: r}
}

func (h *Handler) InitRoute() *gin.Engine {
	router := gin.New()

	router.POST("/tasks", h.CreateTask)
	router.GET("/tasks", h.SelectTasks)
	router.GET("/tasks/:id", h.SelectTaskById)
	router.DELETE("/tasks/:id", h.DeleteTaskById)
	router.GET("/health", h.Health)

	return router
}

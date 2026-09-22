package handlers

import (
	"log"
	"net/http"
	"strconv"

	"main/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateTask(c *gin.Context) {
	var task models.Tasks

	err := c.BindJSON(&task)
	if err != nil {
		newErrResponse(c, http.StatusBadRequest, "Ошибка привязки полученного JSON к полям таблицы tasks")
		return
	}

	taskResponse, err := h.repository.Create(task)
	if err != nil {
		newErrResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	log.Println("Создан task с ID: " + strconv.Itoa(int(taskResponse.Id)) + " и статусом: " + taskResponse.Status)
	c.JSON(http.StatusOK, taskResponse)
}

func (h *Handler) SelectTasks(c *gin.Context) {
	var filter models.TaskFilter

	status := c.Query("status")
	filter.Status = status

	tasks, err := h.repository.Select(filter)
	if err != nil {
		newErrResponse(c, http.StatusBadRequest, "Ошибка выбора записей из таблицы tasks")
		return
	}

	log.Println("Выбраны tasks:", tasks)
	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) SelectTaskById(c *gin.Context) {
	id := c.Param("id")

	log.Println("Выбираем task с ID: " + id)

	taskId, err := strconv.Atoi(id)
	if err != nil {
		newErrResponse(c, http.StatusBadRequest, "Невалидный ID")
	}

	task, err := h.repository.SelectById(taskId)
	if err != nil {
		newErrResponse(c, http.StatusBadRequest, "Task с ID "+id+" не был найден")
		return
	}

	log.Println("Выбран task с ID: " + id + " и статусом: " + task.Status)
	c.JSON(http.StatusOK, task)
}

func (h *Handler) DeleteTaskById(c *gin.Context) {
	id := c.Param("id")

	log.Println("Удаление task с ID: " + id)

	taskId, err := strconv.Atoi(id)

	if err != nil {
		newErrResponse(c, http.StatusBadRequest, "Невалидный ID")
	}

	err = h.repository.DeleteById(taskId)
	if err != nil {
		newErrResponse(c, http.StatusInternalServerError, "Ошибка удаления записи "+id)
		return
	}

	log.Println("Task удалён: " + id)
	c.JSON(http.StatusOK, gin.H{
		"message": "Task удалён: " + id,
	})
}

func (h *Handler) Health(c *gin.Context) {
	err := h.repository.Health()

	if err != nil {
		log.Println("Ошибка подключения к базе данных: " + err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Ошибка подключения к базе данных: " + err.Error(),
		})
	} else {
		log.Println("Подключение к базе данных успешно")
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	}
}

type errResponse struct {
	Message string `json:"message"`
}

type statusResponse struct {
	Status string `json:"status"`
}

func newErrResponse(c *gin.Context, statusCode int, message string) {
	log.Println(message)
	c.AbortWithStatusJSON(statusCode, errResponse{message})
}

package repository

import (
	"fmt"
	"main/models"
	"net/http"
	"unicode/utf8"

	"time"
)

func (r *Repository) Create(taskNew models.Tasks) (models.Tasks, error, int) {
	var task models.Tasks

	task.Title = taskNew.Title
	task.Status = models.StatusNew.String()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	err, errCode := ValidateTask(task, true)
	if err != nil {
		return task, err, errCode
	}

	transaction := r.worker.Db.Begin()

	defer func() {
		if r := recover(); r != nil {
			transaction.Rollback()
		}
	}()

	err = transaction.Create(&task).Error
	if err != nil {
		transaction.Rollback()
		return task, err, http.StatusInternalServerError
	}

	err = transaction.Commit().Error
	if err != nil {
		return task, err, http.StatusInternalServerError
	}

	fullTask, err := r.getFullTask(task.Id)
	if err != nil {
		return task, err, http.StatusInternalServerError
	}

	r.worker.TaskChan <- fullTask

	return fullTask, nil, http.StatusOK
}

func (r *Repository) Select(filter models.TaskFilter) ([]models.Tasks, error, int) {
	var tasks []models.Tasks

	query := r.worker.Db.Model(&tasks)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	err := query.Find(&tasks).Error
	if err != nil {
		return tasks, err, http.StatusNotFound
	}

	return tasks, nil, http.StatusOK
}

func (r *Repository) SelectById(id uint64) (models.Tasks, error, int) {
	var task models.Tasks

	task.Id = id

	fullTask, err := r.getFullTask(task.Id)
	if err != nil {
		return task, err, http.StatusNotFound
	}

	err, errCode := ValidateTask(fullTask, false)
	if err != nil {
		return fullTask, err, errCode
	}

	return fullTask, nil, http.StatusOK
}

func (r *Repository) DeleteById(id uint64) (error, int) {
	_, err := r.getFullTask(id)
	if err != nil {
		return fmt.Errorf("Task с ID %d отсутствует", id), http.StatusNotFound
	}

	result := r.worker.Db.Where("id = ? AND status != ?", id, models.StatusProcessing.String()).Delete(&models.Tasks{}, id)
	if result.Error != nil {
		return err, http.StatusInternalServerError
	} else if result.RowsAffected == 0 {
		return fmt.Errorf("Task с ID %d находится в обработке, удаление невозможно", id), http.StatusConflict
	}

	return nil, http.StatusOK
}

func (r *Repository) Health() (error, int) {
	db, err := r.worker.Db.DB()
	if err != nil {
		return err, http.StatusInternalServerError
	}

	err = db.Ping()
	if err != nil {
		return err, http.StatusInternalServerError
	}

	return nil, http.StatusOK
}

func (r *Repository) getFullTask(id uint64) (models.Tasks, error) {
	var task models.Tasks
	err := r.worker.Db.First(&task, id).Error

	if err != nil {
		return task, err
	}

	return task, nil
}

func (r *Repository) CheckNewTasks() (error, int) {
	var tasks []models.Tasks
	var newCount int = 0

	err := r.worker.Db.Model(&tasks).Find(&tasks).Error
	if err != nil {
		return err, 0
	}

	for _, task := range tasks {
		if task.Status == models.StatusNew.String() || task.Status == models.StatusProcessing.String() {
			r.worker.TaskChan <- task
			newCount++
		}
	}

	return nil, newCount
}

func ValidateTask(task models.Tasks, new bool) (error, int) {
	if task.Id <= 0 && !new {
		return ErrInvalidTaskId, http.StatusBadRequest
	}

	if task.Title == "" {
		return ErrTitleEmpty, http.StatusBadRequest
	}

	if utf8.RuneCountInString(task.Title) > 50 {
		return ErrTitleTooLong, http.StatusBadRequest
	}

	// if task.Status == models.StatusFailed.String() {
	// 	return ErrTaskFailed
	// }

	if task.CreatedAt.Equal((time.Time{})) || task.UpdatedAt.Equal((time.Time{})) {
		return ErrTaskNoTimestamps, http.StatusInternalServerError
	}

	return nil, http.StatusOK
}

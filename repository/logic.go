package repository

import (
	"main/models"

	"time"
)

func (r *Repository) Create(taskNew models.Tasks) (models.Tasks, error) {
	var task models.Tasks

	transaction := r.worker.Db.Begin()

	defer func() {
		if r := recover(); r != nil {
			transaction.Rollback()
		}
	}()

	task.Title = taskNew.Title
	task.Status = models.StatusNew.String()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	err := transaction.Create(&task).Error
	if err != nil {
		transaction.Rollback()
		return task, err
	}

	err = transaction.Commit().Error
	if err != nil {
		return task, err
	}

	fullTask, err := r.getFullTask(task.Id)
	if err != nil {
		return task, err
	}

	err = ValidateTask(fullTask)
	if err != nil {
		return fullTask, err
	}

	r.worker.TaskChan <- fullTask

	return fullTask, nil
}

func (r *Repository) Select(filter models.TaskFilter) ([]models.Tasks, error) {
	var tasks []models.Tasks

	query := r.worker.Db.Model(&tasks)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	err := query.Find(&tasks).Error
	if err != nil {
		return tasks, err
	}

	return tasks, nil
}

func (r *Repository) SelectById(id int) (models.Tasks, error) {
	var task models.Tasks

	task.Id = uint64(id)

	fullTask, err := r.getFullTask(task.Id)
	if err != nil {
		return task, err
	}

	err = ValidateTask(fullTask)
	if err != nil {
		return fullTask, err
	}

	return fullTask, nil
}

func (r *Repository) DeleteById(id int) error {

	err := r.worker.Db.Delete(&models.Tasks{}, id).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Health() error {
	db, err := r.worker.Db.DB()
	if err != nil {
		return err
	}
	return db.Ping()
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
		if task.Status == models.StatusNew.String() {
			r.worker.TaskChan <- task
			newCount++
		}
	}

	return nil, newCount
}

func ValidateTask(task models.Tasks) error {
	if task.Id <= 0 {
		return ErrInvalidTaskId
	}

	// if task.Status == models.StatusFailed.String() {
	// 	return ErrTaskFailed
	// }

	if task.CreatedAt.Equal((time.Time{})) || task.UpdatedAt.Equal((time.Time{})) {
		return ErrTaskNoTimestamps
	}

	return nil
}

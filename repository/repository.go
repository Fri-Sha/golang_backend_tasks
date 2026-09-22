package repository

import (
	"errors"
	"main/worker"
)

type Repository struct {
	worker worker.Worker
}

func NewRepository(worker *worker.Worker) *Repository {
	return &Repository{worker: *worker}
}

var (
	ErrInvalidTaskId = errors.New("Invalid task ID")
	// ErrTaskFailed       = errors.New("Task failed")
	ErrTaskNoTimestamps = errors.New("Task has no timestamps")
)

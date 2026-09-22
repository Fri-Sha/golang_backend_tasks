package worker

import (
	"context"
	"main/models"

	"gorm.io/gorm"
)

const MaxWorkers = 3
const PauseSeconds = 10

type Worker struct {
	Db       *gorm.DB
	TaskChan chan models.Tasks
	ExitChan chan bool
}

func NewWorker(db *gorm.DB) *Worker {
	return &Worker{Db: db, TaskChan: make(chan models.Tasks, MaxWorkers), ExitChan: make(chan bool, MaxWorkers)}
}

func (w *Worker) StartWorker(tasks <-chan models.Tasks, exit chan<- bool, ctx context.Context) {
	for {
		// Если уже была подана команда на завершение работы worker, то не обрабатываем никакие каналы так как следующий task может быть подан в работу
		if ctx.Err() != nil {
			exit <- true
			return
		}

		select {
		case task := <-tasks:
			w.ChangeState(task, models.StatusNew, false)
		case <-ctx.Done():
			exit <- true
			return
		}
	}
}

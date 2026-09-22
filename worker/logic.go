package worker

import (
	"fmt"
	"log"
	"main/models"
	"strings"
	"time"
)

func (w *Worker) ChangeState(task models.Tasks, status models.Status, failedOnce bool) {
	var newStatus models.Status

	log.Println("Обрабатываем task ID: " + fmt.Sprint(task.Id) + " со статусом: " + status.String())

	time.Sleep(time.Second * PauseSeconds)

	transaction := w.Db.Begin()

	if status == models.StatusProcessing && strings.Contains(task.Title, "fail") {
		newStatus = models.StatusFailed
	} else {
		newStatus = models.TransitionStatus(status)
	}

	task.Status = newStatus.String()
	task.UpdatedAt = time.Now()

	err := transaction.Save(&task).Error
	if err != nil {
		log.Println("Ошибка при обновлении task ID: " + fmt.Sprint(task.Id))
		transaction.Rollback()
		return
	}

	err = transaction.Commit().Error
	if err != nil {
		log.Println("Ошибка при коммите task ID: " + fmt.Sprint(task.Id))
		transaction.Rollback()
		return
	}

	if newStatus != models.StatusDone && newStatus != models.StatusFailed {
		w.ChangeState(task, newStatus, false)
	} else if newStatus == models.StatusFailed && !failedOnce {
		log.Println("Повторная обработка task ID: " + fmt.Sprint(task.Id) + ", был получен статус: " + newStatus.String())
		w.ChangeState(task, status, true)
	} else if newStatus == models.StatusFailed && failedOnce {
		log.Println("Закончена обработка task ID: " + fmt.Sprint(task.Id) + ", был получен статус: " + newStatus.String() + ", повторная обработка не удалась")
	} else {
		log.Println("Закончена обработка task ID: " + fmt.Sprint(task.Id))
	}
}

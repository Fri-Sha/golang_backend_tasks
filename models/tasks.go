package models

import (
	"time"
)

// Enum для статусов (подаём ID, получаем текст статуса)
type Status int

const (
	StatusNew Status = iota
	StatusProcessing
	StatusDone
	StatusFailed
)

var statusName = map[Status]string{
	StatusNew:        StatusNewString,
	StatusProcessing: StatusProcessingString,
	StatusDone:       StatusDoneString,
	StatusFailed:     StatusFailedString,
}

func (s Status) String() string {
	return statusName[s]
}

func TransitionStatus(s Status) Status {
	switch s {
	case StatusNew:
		return StatusProcessing
	case StatusProcessing:
		return StatusDone
	default:
		return StatusFailed
	}
}

// Map для получения ID из статуса (подаём текст статуса, получаем его ID)
const (
	StatusNewString        string = "new"
	StatusProcessingString string = "processing"
	StatusDoneString       string = "done"
	StatusFailedString     string = "failed"
)

var statusNameString = map[string]Status{
	StatusNewString:        StatusNew,
	StatusProcessingString: StatusProcessing,
	StatusDoneString:       StatusDone,
	StatusFailedString:     StatusFailed,
}

func GetStatusId(s string) Status {
	return statusNameString[s]
}

// Map для проверки если текущий ID из таблицы tasks обрабатывается
var IsTaskInProcess map[uint64]bool

type Tasks struct {
	Id        uint64    `json:"id" gorm:"column:id;primary_key;auto_increment"`
	Title     string    `json:"title" gorm:"column:title"`
	Status    string    `json:"status" gorm:"column:status"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

type TaskFilter struct {
	Status string `form:"status"`
}

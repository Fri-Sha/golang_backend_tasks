package models

import (
	"time"
)

type Status int

const (
	StatusNew Status = iota
	StatusProcessing
	StatusDone
	StatusFailed
)

var statusName = map[Status]string{
	StatusNew:        "new",
	StatusProcessing: "processing",
	StatusDone:       "done",
	StatusFailed:     "failed",
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

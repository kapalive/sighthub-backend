package tasks

import "time"

type Task struct {
	IDTask       int64      `gorm:"column:id_task;primaryKey;autoIncrement" json:"id_task"`
	Title        string     `gorm:"column:title;size:255;not null"          json:"title"`
	Description  *string    `gorm:"column:description;type:text"            json:"description,omitempty"`
	Status       string     `gorm:"column:status;size:50;default:open"      json:"status"`
	Priority     string     `gorm:"column:priority;size:20;default:medium"  json:"priority"`
	DueDate      *time.Time `gorm:"column:due_date;type:timestamptz"        json:"due_date,omitempty"`
	CreateAt     *time.Time `gorm:"column:create_at;type:timestamptz"       json:"create_at,omitempty"`
	CreatorID    int        `gorm:"column:creator_id;not null"              json:"creator_id"`
	EmployeeID   *int       `gorm:"column:employee_id"                      json:"employee_id,omitempty"`
	ParentTaskID *int64     `gorm:"column:parent_task_id"                   json:"parent_task_id,omitempty"`
}

func (Task) TableName() string { return "tasks" }

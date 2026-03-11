package tasks

import "time"

type TaskComment struct {
	IDTaskComment int64      `gorm:"column:id_task_comment;primaryKey;autoIncrement" json:"id_task_comment"`
	TaskID        int64      `gorm:"column:task_id;not null"                         json:"task_id"`
	AuthorID      int        `gorm:"column:author_id;not null"                       json:"author_id"`
	Message       string     `gorm:"column:message;type:text;not null"               json:"message"`
	CreatedAt     *time.Time `gorm:"column:created_at;type:timestamptz"              json:"created_at,omitempty"`
}

func (TaskComment) TableName() string { return "task_comment" }

package tasks

import "time"

type TaskHistory struct {
	IDTaskHistory      int64      `gorm:"column:id_task_history;primaryKey;autoIncrement" json:"id_task_history"`
	TaskID             int64      `gorm:"column:task_id;not null"                         json:"task_id"`
	OldEmployeeID      *int       `gorm:"column:old_employee_id"                          json:"old_employee_id,omitempty"`
	NewEmployeeID      *int       `gorm:"column:new_employee_id"                          json:"new_employee_id,omitempty"`
	Description        *string    `gorm:"column:description;type:text"                    json:"description,omitempty"`
	LastUpdate         *time.Time `gorm:"column:last_update;type:timestamptz"             json:"last_update,omitempty"`
	OldTaskResourceID  *int64     `gorm:"column:old_task_resource_id"                     json:"old_task_resource_id,omitempty"`
	NewTaskResourceID  *int64     `gorm:"column:new_task_resource_id"                     json:"new_task_resource_id,omitempty"`
	OldStatus          *string    `gorm:"column:old_status;size:50"                       json:"old_status,omitempty"`
	NewStatus          *string    `gorm:"column:new_status;size:50"                       json:"new_status,omitempty"`
}

func (TaskHistory) TableName() string { return "task_history" }

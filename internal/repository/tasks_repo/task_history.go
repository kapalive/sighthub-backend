package tasks_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/tasks"
)

type TaskHistoryRepo struct{ DB *gorm.DB }

func NewTaskHistoryRepo(db *gorm.DB) *TaskHistoryRepo { return &TaskHistoryRepo{DB: db} }

func (r *TaskHistoryRepo) GetByTaskID(taskID int64) ([]tasks.TaskHistory, error) {
	var items []tasks.TaskHistory
	return items, r.DB.Where("task_id = ?", taskID).Order("last_update DESC").Find(&items).Error
}

func (r *TaskHistoryRepo) Create(item *tasks.TaskHistory) error {
	return r.DB.Create(item).Error
}

// Log is a convenience helper that creates a history entry.
func (r *TaskHistoryRepo) Log(taskID int64, oldStatus, newStatus *string, oldEmpID, newEmpID *int, desc *string) error {
	return r.Create(&tasks.TaskHistory{
		TaskID:        taskID,
		OldEmployeeID: oldEmpID,
		NewEmployeeID: newEmpID,
		Description:   desc,
		OldStatus:     oldStatus,
		NewStatus:     newStatus,
	})
}

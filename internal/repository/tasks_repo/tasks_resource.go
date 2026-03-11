package tasks_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/tasks"
)

type TasksResourceRepo struct{ DB *gorm.DB }

func NewTasksResourceRepo(db *gorm.DB) *TasksResourceRepo { return &TasksResourceRepo{DB: db} }

func (r *TasksResourceRepo) GetByID(id int64) (*tasks.TasksResource, error) {
	var item tasks.TasksResource
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *TasksResourceRepo) GetByTaskID(taskID int64) ([]tasks.TasksResource, error) {
	var items []tasks.TasksResource
	return items, r.DB.Where("task_id = ?", taskID).Find(&items).Error
}

func (r *TasksResourceRepo) Create(item *tasks.TasksResource) error {
	return r.DB.Create(item).Error
}

func (r *TasksResourceRepo) Delete(id int64) error {
	return r.DB.Delete(&tasks.TasksResource{}, id).Error
}

func (r *TasksResourceRepo) DeleteByTaskID(taskID int64) error {
	return r.DB.Where("task_id = ?", taskID).Delete(&tasks.TasksResource{}).Error
}

package tasks_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/tasks"
)

type TaskRepo struct{ DB *gorm.DB }

func NewTaskRepo(db *gorm.DB) *TaskRepo { return &TaskRepo{DB: db} }

func (r *TaskRepo) GetByID(id int64) (*tasks.Task, error) {
	var item tasks.Task
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *TaskRepo) GetByAssignee(employeeID int) ([]tasks.Task, error) {
	var items []tasks.Task
	return items, r.DB.Where("employee_id = ?", employeeID).Order("create_at DESC").Find(&items).Error
}

func (r *TaskRepo) GetByCreator(creatorID int) ([]tasks.Task, error) {
	var items []tasks.Task
	return items, r.DB.Where("creator_id = ?", creatorID).Order("create_at DESC").Find(&items).Error
}

func (r *TaskRepo) GetByParent(parentTaskID int64) ([]tasks.Task, error) {
	var items []tasks.Task
	return items, r.DB.Where("parent_task_id = ?", parentTaskID).Find(&items).Error
}

func (r *TaskRepo) GetByStatus(status string) ([]tasks.Task, error) {
	var items []tasks.Task
	return items, r.DB.Where("status = ?", status).Order("priority, due_date").Find(&items).Error
}

func (r *TaskRepo) Create(item *tasks.Task) error {
	return r.DB.Create(item).Error
}

func (r *TaskRepo) Save(item *tasks.Task) error {
	return r.DB.Save(item).Error
}

func (r *TaskRepo) UpdateStatus(id int64, status string) error {
	return r.DB.Model(&tasks.Task{}).Where("id_task = ?", id).Update("status", status).Error
}

func (r *TaskRepo) Delete(id int64) error {
	return r.DB.Delete(&tasks.Task{}, id).Error
}

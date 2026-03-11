package tasks_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/tasks"
)

type TaskResourceRelationRepo struct{ DB *gorm.DB }

func NewTaskResourceRelationRepo(db *gorm.DB) *TaskResourceRelationRepo {
	return &TaskResourceRelationRepo{DB: db}
}

func (r *TaskResourceRelationRepo) GetByID(id int64) (*tasks.TaskResourceRelation, error) {
	var item tasks.TaskResourceRelation
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *TaskResourceRelationRepo) GetByTaskID(taskID int64) ([]tasks.TaskResourceRelation, error) {
	var items []tasks.TaskResourceRelation
	return items, r.DB.Where("task_id = ?", taskID).Find(&items).Error
}

func (r *TaskResourceRelationRepo) Create(item *tasks.TaskResourceRelation) error {
	return r.DB.Create(item).Error
}

func (r *TaskResourceRelationRepo) Delete(id int64) error {
	return r.DB.Delete(&tasks.TaskResourceRelation{}, id).Error
}

func (r *TaskResourceRelationRepo) DeleteByTaskID(taskID int64) error {
	return r.DB.Where("task_id = ?", taskID).Delete(&tasks.TaskResourceRelation{}).Error
}

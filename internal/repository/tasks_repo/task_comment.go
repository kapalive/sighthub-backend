package tasks_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/tasks"
)

type TaskCommentRepo struct{ DB *gorm.DB }

func NewTaskCommentRepo(db *gorm.DB) *TaskCommentRepo { return &TaskCommentRepo{DB: db} }

func (r *TaskCommentRepo) GetByID(id int64) (*tasks.TaskComment, error) {
	var item tasks.TaskComment
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *TaskCommentRepo) GetByTaskID(taskID int64) ([]tasks.TaskComment, error) {
	var items []tasks.TaskComment
	return items, r.DB.Where("task_id = ?", taskID).Order("created_at").Find(&items).Error
}

func (r *TaskCommentRepo) Create(item *tasks.TaskComment) error {
	return r.DB.Create(item).Error
}

func (r *TaskCommentRepo) Save(item *tasks.TaskComment) error {
	return r.DB.Save(item).Error
}

func (r *TaskCommentRepo) Delete(id int64) error {
	return r.DB.Delete(&tasks.TaskComment{}, id).Error
}

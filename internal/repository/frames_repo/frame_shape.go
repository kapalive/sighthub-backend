package frames_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/frames"
)

type FrameShapeRepo struct{ DB *gorm.DB }

func NewFrameShapeRepo(db *gorm.DB) *FrameShapeRepo { return &FrameShapeRepo{DB: db} }

func (r *FrameShapeRepo) GetAll() ([]frames.FrameShape, error) {
	var items []frames.FrameShape
	return items, r.DB.Find(&items).Error
}

func (r *FrameShapeRepo) GetByID(id int) (*frames.FrameShape, error) {
	var v frames.FrameShape
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *FrameShapeRepo) Create(v *frames.FrameShape) error {
	return r.DB.Create(v).Error
}

func (r *FrameShapeRepo) Save(v *frames.FrameShape) error {
	return r.DB.Save(v).Error
}

func (r *FrameShapeRepo) Delete(id int) error {
	return r.DB.Delete(&frames.FrameShape{}, id).Error
}

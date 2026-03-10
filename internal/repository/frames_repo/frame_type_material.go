package frames_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/frames"
)

type FrameTypeMaterialRepo struct{ DB *gorm.DB }

func NewFrameTypeMaterialRepo(db *gorm.DB) *FrameTypeMaterialRepo {
	return &FrameTypeMaterialRepo{DB: db}
}

func (r *FrameTypeMaterialRepo) GetAll() ([]frames.FrameTypeMaterial, error) {
	var items []frames.FrameTypeMaterial
	return items, r.DB.Find(&items).Error
}

func (r *FrameTypeMaterialRepo) GetByID(id int) (*frames.FrameTypeMaterial, error) {
	var v frames.FrameTypeMaterial
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *FrameTypeMaterialRepo) Create(v *frames.FrameTypeMaterial) error {
	return r.DB.Create(v).Error
}

func (r *FrameTypeMaterialRepo) Save(v *frames.FrameTypeMaterial) error {
	return r.DB.Save(v).Error
}

func (r *FrameTypeMaterialRepo) Delete(id int) error {
	return r.DB.Delete(&frames.FrameTypeMaterial{}, id).Error
}

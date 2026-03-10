package external_sle

import (
	"errors"

	"gorm.io/gorm"
	e "sighthub-backend/internal/models/medical/vision_exam/external_sle"
)

type VisualFieldsRepo struct{ DB *gorm.DB }

func NewVisualFieldsRepo(db *gorm.DB) *VisualFieldsRepo {
	return &VisualFieldsRepo{DB: db}
}

func (r *VisualFieldsRepo) GetByID(id int64) (*e.VisualFields, error) {
	var v e.VisualFields
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *VisualFieldsRepo) Create(v *e.VisualFields) error {
	return r.DB.Create(v).Error
}

func (r *VisualFieldsRepo) Save(v *e.VisualFields) error {
	return r.DB.Save(v).Error
}

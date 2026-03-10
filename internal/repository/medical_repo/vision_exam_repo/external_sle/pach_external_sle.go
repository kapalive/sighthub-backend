package external_sle

import (
	"errors"

	"gorm.io/gorm"
	e "sighthub-backend/internal/models/medical/vision_exam/external_sle"
)

type PachExternalSleRepo struct{ DB *gorm.DB }

func NewPachExternalSleRepo(db *gorm.DB) *PachExternalSleRepo {
	return &PachExternalSleRepo{DB: db}
}

func (r *PachExternalSleRepo) GetByID(id int64) (*e.PachExternalSle, error) {
	var v e.PachExternalSle
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *PachExternalSleRepo) Create(v *e.PachExternalSle) error {
	return r.DB.Create(v).Error
}

func (r *PachExternalSleRepo) Save(v *e.PachExternalSle) error {
	return r.DB.Save(v).Error
}

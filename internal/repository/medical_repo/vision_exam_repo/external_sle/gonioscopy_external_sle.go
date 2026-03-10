package external_sle

import (
	"errors"

	"gorm.io/gorm"
	e "sighthub-backend/internal/models/medical/vision_exam/external_sle"
)

type GonioscopyExternalSleRepo struct{ DB *gorm.DB }

func NewGonioscopyExternalSleRepo(db *gorm.DB) *GonioscopyExternalSleRepo {
	return &GonioscopyExternalSleRepo{DB: db}
}

func (r *GonioscopyExternalSleRepo) GetByID(id int64) (*e.GonioscopyExternalSle, error) {
	var v e.GonioscopyExternalSle
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *GonioscopyExternalSleRepo) Create(v *e.GonioscopyExternalSle) error {
	return r.DB.Create(v).Error
}

func (r *GonioscopyExternalSleRepo) Save(v *e.GonioscopyExternalSle) error {
	return r.DB.Save(v).Error
}

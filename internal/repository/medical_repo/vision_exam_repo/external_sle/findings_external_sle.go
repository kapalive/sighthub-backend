package external_sle

import (
	"errors"

	"gorm.io/gorm"
	e "sighthub-backend/internal/models/medical/vision_exam/external_sle"
)

type FindingsExternalSleRepo struct{ DB *gorm.DB }

func NewFindingsExternalSleRepo(db *gorm.DB) *FindingsExternalSleRepo {
	return &FindingsExternalSleRepo{DB: db}
}

func (r *FindingsExternalSleRepo) GetByID(id int64) (*e.FindingsExternalSle, error) {
	var v e.FindingsExternalSle
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *FindingsExternalSleRepo) Create(v *e.FindingsExternalSle) error {
	return r.DB.Create(v).Error
}

func (r *FindingsExternalSleRepo) Save(v *e.FindingsExternalSle) error {
	return r.DB.Save(v).Error
}

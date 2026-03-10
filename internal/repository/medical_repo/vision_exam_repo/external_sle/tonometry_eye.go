package external_sle

import (
	"errors"

	"gorm.io/gorm"
	e "sighthub-backend/internal/models/medical/vision_exam/external_sle"
)

type TonometryEyeRepo struct{ DB *gorm.DB }

func NewTonometryEyeRepo(db *gorm.DB) *TonometryEyeRepo {
	return &TonometryEyeRepo{DB: db}
}

func (r *TonometryEyeRepo) GetByID(id int64) (*e.TonometryEye, error) {
	var v e.TonometryEye
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *TonometryEyeRepo) Create(v *e.TonometryEye) error {
	return r.DB.Create(v).Error
}

func (r *TonometryEyeRepo) Save(v *e.TonometryEye) error {
	return r.DB.Save(v).Error
}

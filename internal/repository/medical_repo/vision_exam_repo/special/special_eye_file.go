package special

import (
	"errors"

	"gorm.io/gorm"
	sp "sighthub-backend/internal/models/medical/vision_exam/special"
)

type SpecialEyeFileRepo struct{ DB *gorm.DB }

func NewSpecialEyeFileRepo(db *gorm.DB) *SpecialEyeFileRepo {
	return &SpecialEyeFileRepo{DB: db}
}

func (r *SpecialEyeFileRepo) GetByID(id int64) (*sp.SpecialEyeFile, error) {
	var v sp.SpecialEyeFile
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *SpecialEyeFileRepo) Create(v *sp.SpecialEyeFile) error {
	return r.DB.Create(v).Error
}

func (r *SpecialEyeFileRepo) Save(v *sp.SpecialEyeFile) error {
	return r.DB.Save(v).Error
}

func (r *SpecialEyeFileRepo) Delete(id int64) error {
	return r.DB.Delete(&sp.SpecialEyeFile{}, id).Error
}

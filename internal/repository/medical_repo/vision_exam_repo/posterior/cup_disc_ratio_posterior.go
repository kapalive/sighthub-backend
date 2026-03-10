package posterior

import (
	"errors"

	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/posterior"
)

type CupDiscRatioPosteriorRepo struct{ DB *gorm.DB }

func NewCupDiscRatioPosteriorRepo(db *gorm.DB) *CupDiscRatioPosteriorRepo {
	return &CupDiscRatioPosteriorRepo{DB: db}
}

func (r *CupDiscRatioPosteriorRepo) GetByID(id int64) (*p.CupDiscRatioPosterior, error) {
	var v p.CupDiscRatioPosterior
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *CupDiscRatioPosteriorRepo) Create(v *p.CupDiscRatioPosterior) error {
	return r.DB.Create(v).Error
}

func (r *CupDiscRatioPosteriorRepo) Save(v *p.CupDiscRatioPosterior) error {
	return r.DB.Save(v).Error
}

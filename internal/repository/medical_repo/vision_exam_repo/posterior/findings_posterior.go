package posterior

import (
	"errors"

	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/posterior"
)

type FindingsPosteriorRepo struct{ DB *gorm.DB }

func NewFindingsPosteriorRepo(db *gorm.DB) *FindingsPosteriorRepo {
	return &FindingsPosteriorRepo{DB: db}
}

func (r *FindingsPosteriorRepo) GetByID(id int64) (*p.FindingsPosterior, error) {
	var v p.FindingsPosterior
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *FindingsPosteriorRepo) Create(v *p.FindingsPosterior) error {
	return r.DB.Create(v).Error
}

func (r *FindingsPosteriorRepo) Save(v *p.FindingsPosterior) error {
	return r.DB.Save(v).Error
}

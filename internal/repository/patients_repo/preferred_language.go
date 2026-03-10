package patients_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/patients"
)

type PreferredLanguageRepo struct{ DB *gorm.DB }

func NewPreferredLanguageRepo(db *gorm.DB) *PreferredLanguageRepo {
	return &PreferredLanguageRepo{DB: db}
}

func (r *PreferredLanguageRepo) GetAll() ([]patients.PreferredLanguage, error) {
	var items []patients.PreferredLanguage
	if err := r.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PreferredLanguageRepo) GetByID(id int) (*patients.PreferredLanguage, error) {
	var item patients.PreferredLanguage
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

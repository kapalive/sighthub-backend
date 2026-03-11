package service_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/service"
)

type ProfessionalServiceScopeRepo struct{ DB *gorm.DB }

func NewProfessionalServiceScopeRepo(db *gorm.DB) *ProfessionalServiceScopeRepo {
	return &ProfessionalServiceScopeRepo{DB: db}
}

func (r *ProfessionalServiceScopeRepo) GetAll() ([]service.ProfessionalServiceScope, error) {
	var items []service.ProfessionalServiceScope
	return items, r.DB.Find(&items).Error
}

func (r *ProfessionalServiceScopeRepo) GetByID(id int) (*service.ProfessionalServiceScope, error) {
	var item service.ProfessionalServiceScope
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

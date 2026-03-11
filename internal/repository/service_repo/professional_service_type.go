package service_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/service"
)

type ProfessionalServiceTypeRepo struct{ DB *gorm.DB }

func NewProfessionalServiceTypeRepo(db *gorm.DB) *ProfessionalServiceTypeRepo {
	return &ProfessionalServiceTypeRepo{DB: db}
}

func (r *ProfessionalServiceTypeRepo) GetAll() ([]service.ProfessionalServiceType, error) {
	var items []service.ProfessionalServiceType
	return items, r.DB.Find(&items).Error
}

func (r *ProfessionalServiceTypeRepo) GetByID(id int) (*service.ProfessionalServiceType, error) {
	var item service.ProfessionalServiceType
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

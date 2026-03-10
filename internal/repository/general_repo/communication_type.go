package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type CommunicationTypeRepo struct{ DB *gorm.DB }

func NewCommunicationTypeRepo(db *gorm.DB) *CommunicationTypeRepo {
	return &CommunicationTypeRepo{DB: db}
}

func (r *CommunicationTypeRepo) GetAll() ([]general.CommunicationType, error) {
	var items []general.CommunicationType
	return items, r.DB.Find(&items).Error
}

func (r *CommunicationTypeRepo) GetByID(id int) (*general.CommunicationType, error) {
	var v general.CommunicationType
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

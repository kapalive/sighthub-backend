// internal/repository/lab_ticket_repo/lab_ticket_contact_lens_services.go
package lab_ticket_repo

import (
	"errors"
	"gorm.io/gorm"
	lt "sighthub-backend/internal/models/lab_ticket"
)

type ContactLensServicesRepo struct{ DB *gorm.DB }

func NewContactLensServicesRepo(db *gorm.DB) *ContactLensServicesRepo {
	return &ContactLensServicesRepo{DB: db}
}

func (r *ContactLensServicesRepo) GetAll() ([]lt.LabTicketContactLensService, error) {
	var rows []lt.LabTicketContactLensService
	return rows, r.DB.Find(&rows).Error
}

func (r *ContactLensServicesRepo) GetByID(id int64) (*lt.LabTicketContactLensService, error) {
	var row lt.LabTicketContactLensService
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

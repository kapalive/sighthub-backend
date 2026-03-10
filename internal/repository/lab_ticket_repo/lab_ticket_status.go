// internal/repository/lab_ticket_repo/lab_ticket_status.go
package lab_ticket_repo

import (
	"errors"
	"gorm.io/gorm"
	lt "sighthub-backend/internal/models/lab_ticket"
)

type LabTicketStatusRepo struct{ DB *gorm.DB }

func NewLabTicketStatusRepo(db *gorm.DB) *LabTicketStatusRepo {
	return &LabTicketStatusRepo{DB: db}
}

func (r *LabTicketStatusRepo) GetAll() ([]lt.LabTicketStatus, error) {
	var rows []lt.LabTicketStatus
	return rows, r.DB.Find(&rows).Error
}

func (r *LabTicketStatusRepo) GetByID(id int64) (*lt.LabTicketStatus, error) {
	var row lt.LabTicketStatus
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

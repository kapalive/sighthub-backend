// internal/repository/lab_ticket_repo/lab_ticket_frame.go
package lab_ticket_repo

import (
	"errors"
	"gorm.io/gorm"
	lt "sighthub-backend/internal/models/lab_ticket"
)

type LabTicketFrameRepo struct{ DB *gorm.DB }

func NewLabTicketFrameRepo(db *gorm.DB) *LabTicketFrameRepo { return &LabTicketFrameRepo{DB: db} }

func (r *LabTicketFrameRepo) GetByID(id int64) (*lt.LabTicketFrame, error) {
	var row lt.LabTicketFrame
	err := r.DB.
		Preload("FrameTypeMaterial").
		Preload("FrameShape").
		First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Create создаёт спецификацию оправы.
func (r *LabTicketFrameRepo) Create(f *lt.LabTicketFrame) error {
	return r.DB.Create(f).Error
}

// Save сохраняет (создаёт или обновляет).
func (r *LabTicketFrameRepo) Save(f *lt.LabTicketFrame) error {
	return r.DB.Save(f).Error
}

func (r *LabTicketFrameRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

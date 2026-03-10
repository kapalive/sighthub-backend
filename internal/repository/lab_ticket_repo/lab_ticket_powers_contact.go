// internal/repository/lab_ticket_repo/lab_ticket_powers_contact.go
package lab_ticket_repo

import (
	"errors"
	"gorm.io/gorm"
	lt "sighthub-backend/internal/models/lab_ticket"
)

type LabTicketPowersContactRepo struct{ DB *gorm.DB }

func NewLabTicketPowersContactRepo(db *gorm.DB) *LabTicketPowersContactRepo {
	return &LabTicketPowersContactRepo{DB: db}
}

func (r *LabTicketPowersContactRepo) GetByID(id int64) (*lt.LabTicketPowersContact, error) {
	var row lt.LabTicketPowersContact
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Create создаёт запись с рецептурными данными (контактные линзы).
func (r *LabTicketPowersContactRepo) Create(p *lt.LabTicketPowersContact) error {
	return r.DB.Create(p).Error
}

// Save сохраняет (создаёт или обновляет) запись.
func (r *LabTicketPowersContactRepo) Save(p *lt.LabTicketPowersContact) error {
	return r.DB.Save(p).Error
}

func (r *LabTicketPowersContactRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

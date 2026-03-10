// internal/repository/lab_ticket_repo/lab_ticket_contact.go
package lab_ticket_repo

import (
	"errors"
	"gorm.io/gorm"
	lt "sighthub-backend/internal/models/lab_ticket"
)

type LabTicketContactRepo struct{ DB *gorm.DB }

func NewLabTicketContactRepo(db *gorm.DB) *LabTicketContactRepo {
	return &LabTicketContactRepo{DB: db}
}

func (r *LabTicketContactRepo) GetByID(id int64) (*lt.LabTicketContact, error) {
	var row lt.LabTicketContact
	err := r.DB.
		Preload("ContactLensService").
		Preload("BrandContactLens").
		Preload("Manufacturer").
		First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Create создаёт контактную информацию тикета.
func (r *LabTicketContactRepo) Create(c *lt.LabTicketContact) error {
	return r.DB.Create(c).Error
}

// Save сохраняет (создаёт или обновляет).
func (r *LabTicketContactRepo) Save(c *lt.LabTicketContact) error {
	return r.DB.Save(c).Error
}

func (r *LabTicketContactRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

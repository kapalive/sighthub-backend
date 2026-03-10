// internal/repository/lab_ticket_repo/lab_ticket_powers.go
package lab_ticket_repo

import (
	"errors"
	"gorm.io/gorm"
	lt "sighthub-backend/internal/models/lab_ticket"
)

type LabTicketPowersRepo struct{ DB *gorm.DB }

func NewLabTicketPowersRepo(db *gorm.DB) *LabTicketPowersRepo {
	return &LabTicketPowersRepo{DB: db}
}

func (r *LabTicketPowersRepo) GetByID(id int64) (*lt.LabTicketPowers, error) {
	var row lt.LabTicketPowers
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Create создаёт запись с рецептурными данными (очки).
func (r *LabTicketPowersRepo) Create(p *lt.LabTicketPowers) error {
	return r.DB.Create(p).Error
}

// Save сохраняет (создаёт или обновляет) запись powers.
func (r *LabTicketPowersRepo) Save(p *lt.LabTicketPowers) error {
	return r.DB.Save(p).Error
}

func (r *LabTicketPowersRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

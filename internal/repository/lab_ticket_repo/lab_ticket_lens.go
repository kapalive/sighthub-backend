// internal/repository/lab_ticket_repo/lab_ticket_lens.go
package lab_ticket_repo

import (
	"errors"
	"gorm.io/gorm"
	lt "sighthub-backend/internal/models/lab_ticket"
)

type LabTicketLensRepo struct{ DB *gorm.DB }

func NewLabTicketLensRepo(db *gorm.DB) *LabTicketLensRepo { return &LabTicketLensRepo{DB: db} }

func (r *LabTicketLensRepo) GetByID(id int64) (*lt.LabTicketLens, error) {
	var row lt.LabTicketLens
	err := r.DB.
		Preload("LensType").
		Preload("LensesMaterial").
		Preload("LensSafetyThickness").
		Preload("LensEdge").
		Preload("LensTintColor").
		Preload("LensSampleColor").
		First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Create создаёт спецификацию линзы для тикета.
func (r *LabTicketLensRepo) Create(l *lt.LabTicketLens) error {
	return r.DB.Create(l).Error
}

// Save сохраняет (создаёт или обновляет).
func (r *LabTicketLensRepo) Save(l *lt.LabTicketLens) error {
	return r.DB.Save(l).Error
}

func (r *LabTicketLensRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

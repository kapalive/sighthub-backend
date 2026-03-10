package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type RaceRepo struct{ DB *gorm.DB }

func NewRaceRepo(db *gorm.DB) *RaceRepo { return &RaceRepo{DB: db} }

func (r *RaceRepo) GetAll() ([]general.Race, error) {
	var items []general.Race
	return items, r.DB.Find(&items).Error
}

func (r *RaceRepo) GetByID(id int) (*general.Race, error) {
	var v general.Race
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

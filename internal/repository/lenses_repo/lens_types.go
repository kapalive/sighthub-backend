package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensTypeRepo struct{ DB *gorm.DB }

func NewLensTypeRepo(db *gorm.DB) *LensTypeRepo { return &LensTypeRepo{DB: db} }

func (r *LensTypeRepo) GetAll() ([]lenses.LensType, error) {
	var items []lenses.LensType
	return items, r.DB.Find(&items).Error
}

func (r *LensTypeRepo) GetByID(id int) (*lenses.LensType, error) {
	var v lenses.LensType
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensTypeRepo) Create(v *lenses.LensType) error { return r.DB.Create(v).Error }
func (r *LensTypeRepo) Save(v *lenses.LensType) error   { return r.DB.Save(v).Error }
func (r *LensTypeRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.LensType{}, id).Error
}

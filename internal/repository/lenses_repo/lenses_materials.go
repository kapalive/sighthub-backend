package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensesMaterialRepo struct{ DB *gorm.DB }

func NewLensesMaterialRepo(db *gorm.DB) *LensesMaterialRepo { return &LensesMaterialRepo{DB: db} }

func (r *LensesMaterialRepo) GetAll() ([]lenses.LensesMaterial, error) {
	var items []lenses.LensesMaterial
	return items, r.DB.Find(&items).Error
}

func (r *LensesMaterialRepo) GetByID(id int64) (*lenses.LensesMaterial, error) {
	var v lenses.LensesMaterial
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensesMaterialRepo) Create(v *lenses.LensesMaterial) error { return r.DB.Create(v).Error }
func (r *LensesMaterialRepo) Save(v *lenses.LensesMaterial) error   { return r.DB.Save(v).Error }
func (r *LensesMaterialRepo) Delete(id int64) error {
	return r.DB.Delete(&lenses.LensesMaterial{}, id).Error
}

package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensSafetyThicknessRepo struct{ DB *gorm.DB }

func NewLensSafetyThicknessRepo(db *gorm.DB) *LensSafetyThicknessRepo { return &LensSafetyThicknessRepo{DB: db} }

func (r *LensSafetyThicknessRepo) GetAll() ([]lenses.LensSafetyThickness, error) {
	var items []lenses.LensSafetyThickness
	return items, r.DB.Find(&items).Error
}

func (r *LensSafetyThicknessRepo) GetByID(id int) (*lenses.LensSafetyThickness, error) {
	var v lenses.LensSafetyThickness
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensSafetyThicknessRepo) Create(v *lenses.LensSafetyThickness) error { return r.DB.Create(v).Error }
func (r *LensSafetyThicknessRepo) Save(v *lenses.LensSafetyThickness) error   { return r.DB.Save(v).Error }
func (r *LensSafetyThicknessRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.LensSafetyThickness{}, id).Error
}

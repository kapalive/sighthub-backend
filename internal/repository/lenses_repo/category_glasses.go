package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type CategoryGlassesRepo struct{ DB *gorm.DB }

func NewCategoryGlassesRepo(db *gorm.DB) *CategoryGlassesRepo { return &CategoryGlassesRepo{DB: db} }

func (r *CategoryGlassesRepo) GetAll() ([]lenses.CategoryGlasses, error) {
	var items []lenses.CategoryGlasses
	return items, r.DB.Find(&items).Error
}

func (r *CategoryGlassesRepo) GetByID(id int) (*lenses.CategoryGlasses, error) {
	var v lenses.CategoryGlasses
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *CategoryGlassesRepo) Create(v *lenses.CategoryGlasses) error { return r.DB.Create(v).Error }
func (r *CategoryGlassesRepo) Save(v *lenses.CategoryGlasses) error   { return r.DB.Save(v).Error }
func (r *CategoryGlassesRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.CategoryGlasses{}, id).Error
}

package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensEdgeRepo struct{ DB *gorm.DB }

func NewLensEdgeRepo(db *gorm.DB) *LensEdgeRepo { return &LensEdgeRepo{DB: db} }

func (r *LensEdgeRepo) GetAll() ([]lenses.LensEdge, error) {
	var items []lenses.LensEdge
	return items, r.DB.Find(&items).Error
}

func (r *LensEdgeRepo) GetByID(id int) (*lenses.LensEdge, error) {
	var v lenses.LensEdge
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensEdgeRepo) Create(v *lenses.LensEdge) error { return r.DB.Create(v).Error }
func (r *LensEdgeRepo) Save(v *lenses.LensEdge) error   { return r.DB.Save(v).Error }
func (r *LensEdgeRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.LensEdge{}, id).Error
}

package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensTreatmentsRepo struct{ DB *gorm.DB }

func NewLensTreatmentsRepo(db *gorm.DB) *LensTreatmentsRepo { return &LensTreatmentsRepo{DB: db} }

func (r *LensTreatmentsRepo) GetByID(id int64) (*lenses.LensTreatments, error) {
	var v lenses.LensTreatments
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensTreatmentsRepo) GetByVendorID(vendorID int) ([]lenses.LensTreatments, error) {
	var items []lenses.LensTreatments
	return items, r.DB.Where("vendor_id = ? AND can_lookup = true", vendorID).
		Order("item_nbr").Find(&items).Error
}

func (r *LensTreatmentsRepo) Search(query string) ([]lenses.LensTreatments, error) {
	var items []lenses.LensTreatments
	return items, r.DB.
		Where("item_nbr ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%").
		Find(&items).Error
}

func (r *LensTreatmentsRepo) Create(v *lenses.LensTreatments) error { return r.DB.Create(v).Error }
func (r *LensTreatmentsRepo) Save(v *lenses.LensTreatments) error   { return r.DB.Save(v).Error }
func (r *LensTreatmentsRepo) Delete(id int64) error {
	return r.DB.Delete(&lenses.LensTreatments{}, id).Error
}

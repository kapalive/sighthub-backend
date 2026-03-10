package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensesRepo struct{ DB *gorm.DB }

func NewLensesRepo(db *gorm.DB) *LensesRepo { return &LensesRepo{DB: db} }

func (r *LensesRepo) GetByID(id int) (*lenses.Lenses, error) {
	var v lenses.Lenses
	if err := r.DB.
		Preload("LensSeries").
		Preload("LensType").
		Preload("LensesMaterial").
		Preload("Brand").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensesRepo) GetByVendorID(vendorID int) ([]lenses.Lenses, error) {
	var items []lenses.Lenses
	return items, r.DB.
		Preload("LensSeries").Preload("LensType").Preload("LensesMaterial").
		Where("vendor_id = ?", vendorID).Find(&items).Error
}

func (r *LensesRepo) GetByBrandID(brandID int) ([]lenses.Lenses, error) {
	var items []lenses.Lenses
	return items, r.DB.Where("brand_lens_id = ?", brandID).Find(&items).Error
}

func (r *LensesRepo) Search(query string) ([]lenses.Lenses, error) {
	var items []lenses.Lenses
	return items, r.DB.
		Where("lens_name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%").
		Find(&items).Error
}

func (r *LensesRepo) Create(v *lenses.Lenses) error { return r.DB.Create(v).Error }
func (r *LensesRepo) Save(v *lenses.Lenses) error   { return r.DB.Save(v).Error }
func (r *LensesRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.Lenses{}, id).Error
}

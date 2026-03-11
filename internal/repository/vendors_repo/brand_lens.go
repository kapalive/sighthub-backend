package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type BrandLensRepo struct{ DB *gorm.DB }

func NewBrandLensRepo(db *gorm.DB) *BrandLensRepo { return &BrandLensRepo{DB: db} }

func (r *BrandLensRepo) GetAll() ([]vendors.BrandLens, error) {
	var items []vendors.BrandLens
	return items, r.DB.Find(&items).Error
}

func (r *BrandLensRepo) GetByID(id int) (*vendors.BrandLens, error) {
	var item vendors.BrandLens
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *BrandLensRepo) Search(query string) ([]vendors.BrandLens, error) {
	var items []vendors.BrandLens
	q := "%" + query + "%"
	return items, r.DB.Where("brand_name ILIKE ?", q).Find(&items).Error
}

func (r *BrandLensRepo) Create(item *vendors.BrandLens) error {
	return r.DB.Create(item).Error
}

func (r *BrandLensRepo) Save(item *vendors.BrandLens) error {
	return r.DB.Save(item).Error
}

func (r *BrandLensRepo) Delete(id int) error {
	return r.DB.Delete(&vendors.BrandLens{}, id).Error
}

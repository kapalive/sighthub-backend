package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type BrandContactLensRepo struct{ DB *gorm.DB }

func NewBrandContactLensRepo(db *gorm.DB) *BrandContactLensRepo {
	return &BrandContactLensRepo{DB: db}
}

func (r *BrandContactLensRepo) GetAll() ([]vendors.BrandContactLens, error) {
	var items []vendors.BrandContactLens
	return items, r.DB.Find(&items).Error
}

func (r *BrandContactLensRepo) GetByID(id int) (*vendors.BrandContactLens, error) {
	var item vendors.BrandContactLens
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *BrandContactLensRepo) Search(query string) ([]vendors.BrandContactLens, error) {
	var items []vendors.BrandContactLens
	q := "%" + query + "%"
	return items, r.DB.Where("brand_name ILIKE ?", q).Find(&items).Error
}

func (r *BrandContactLensRepo) Create(item *vendors.BrandContactLens) error {
	return r.DB.Create(item).Error
}

func (r *BrandContactLensRepo) Save(item *vendors.BrandContactLens) error {
	return r.DB.Save(item).Error
}

func (r *BrandContactLensRepo) Delete(id int) error {
	return r.DB.Delete(&vendors.BrandContactLens{}, id).Error
}

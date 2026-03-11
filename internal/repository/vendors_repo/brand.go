package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type BrandRepo struct{ DB *gorm.DB }

func NewBrandRepo(db *gorm.DB) *BrandRepo { return &BrandRepo{DB: db} }

func (r *BrandRepo) GetAll() ([]vendors.Brand, error) {
	var items []vendors.Brand
	return items, r.DB.Find(&items).Error
}

func (r *BrandRepo) GetByID(id int) (*vendors.Brand, error) {
	var item vendors.Brand
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *BrandRepo) Search(query string) ([]vendors.Brand, error) {
	var items []vendors.Brand
	q := "%" + query + "%"
	return items, r.DB.Where("brand_name ILIKE ?", q).Find(&items).Error
}

func (r *BrandRepo) Create(item *vendors.Brand) error {
	return r.DB.Create(item).Error
}

func (r *BrandRepo) Save(item *vendors.Brand) error {
	return r.DB.Save(item).Error
}

func (r *BrandRepo) Delete(id int) error {
	return r.DB.Delete(&vendors.Brand{}, id).Error
}

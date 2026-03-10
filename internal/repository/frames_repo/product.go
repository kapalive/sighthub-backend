package frames_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/frames"
)

type ProductRepo struct{ DB *gorm.DB }

func NewProductRepo(db *gorm.DB) *ProductRepo { return &ProductRepo{DB: db} }

func (r *ProductRepo) GetByID(id int64) (*frames.Product, error) {
	var v frames.Product
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ProductRepo) GetByVendorID(vendorID int64) ([]frames.Product, error) {
	var items []frames.Product
	return items, r.DB.Where("vendor_id = ?", vendorID).Find(&items).Error
}

func (r *ProductRepo) GetByBrandID(brandID int64) ([]frames.Product, error) {
	var items []frames.Product
	return items, r.DB.Where("brand_id = ?", brandID).Find(&items).Error
}

func (r *ProductRepo) Search(query string) ([]frames.Product, error) {
	var items []frames.Product
	return items, r.DB.
		Where("title_product ILIKE ?", "%"+query+"%").
		Find(&items).Error
}

func (r *ProductRepo) Create(v *frames.Product) error {
	return r.DB.Create(v).Error
}

func (r *ProductRepo) Save(v *frames.Product) error {
	return r.DB.Save(v).Error
}

func (r *ProductRepo) Delete(id int64) error {
	return r.DB.Delete(&frames.Product{}, id).Error
}

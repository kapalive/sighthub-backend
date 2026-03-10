package other_products_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/other_products"
)

type CrossSellProductRepo struct{ DB *gorm.DB }

func NewCrossSellProductRepo(db *gorm.DB) *CrossSellProductRepo {
	return &CrossSellProductRepo{DB: db}
}

func (r *CrossSellProductRepo) GetByID(id int64) (*other_products.CrossSellProduct, error) {
	var item other_products.CrossSellProduct
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *CrossSellProductRepo) GetAll() ([]other_products.CrossSellProduct, error) {
	var items []other_products.CrossSellProduct
	if err := r.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CrossSellProductRepo) GetByVendorID(vendorID int64) ([]other_products.CrossSellProduct, error) {
	var items []other_products.CrossSellProduct
	if err := r.DB.Where("vendor_id = ?", vendorID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CrossSellProductRepo) GetByBrandID(brandID int64) ([]other_products.CrossSellProduct, error) {
	var items []other_products.CrossSellProduct
	if err := r.DB.Where("brand_id = ?", brandID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CrossSellProductRepo) Search(query string) ([]other_products.CrossSellProduct, error) {
	var items []other_products.CrossSellProduct
	q := "%" + query + "%"
	if err := r.DB.Where("title ILIKE ?", q).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *CrossSellProductRepo) Create(item *other_products.CrossSellProduct) error {
	return r.DB.Create(item).Error
}

func (r *CrossSellProductRepo) Save(item *other_products.CrossSellProduct) error {
	return r.DB.Save(item).Error
}

func (r *CrossSellProductRepo) Delete(id int64) error {
	return r.DB.Delete(&other_products.CrossSellProduct{}, id).Error
}

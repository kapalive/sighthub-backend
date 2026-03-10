package other_products_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/other_products"
)

type VariantCrSlProductRepo struct{ DB *gorm.DB }

func NewVariantCrSlProductRepo(db *gorm.DB) *VariantCrSlProductRepo {
	return &VariantCrSlProductRepo{DB: db}
}

func (r *VariantCrSlProductRepo) GetByID(id int64) (*other_products.VariantCrSlProduct, error) {
	var item other_products.VariantCrSlProduct
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *VariantCrSlProductRepo) GetByCrossSellProductID(crossSellProductID int64) ([]other_products.VariantCrSlProduct, error) {
	var items []other_products.VariantCrSlProduct
	if err := r.DB.Where("cross_sell_products_id = ?", crossSellProductID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *VariantCrSlProductRepo) Create(item *other_products.VariantCrSlProduct) error {
	return r.DB.Create(item).Error
}

func (r *VariantCrSlProductRepo) Save(item *other_products.VariantCrSlProduct) error {
	return r.DB.Save(item).Error
}

func (r *VariantCrSlProductRepo) Delete(id int64) error {
	return r.DB.Delete(&other_products.VariantCrSlProduct{}, id).Error
}

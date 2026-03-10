package frames_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/frames"
)

type ModelRepo struct{ DB *gorm.DB }

func NewModelRepo(db *gorm.DB) *ModelRepo { return &ModelRepo{DB: db} }

func (r *ModelRepo) GetByID(id int64) (*frames.Model, error) {
	var v frames.Model
	if err := r.DB.
		Preload("Product").
		Preload("CategoryGlasses").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ModelRepo) GetByProductID(productID int64) ([]frames.Model, error) {
	var items []frames.Model
	return items, r.DB.Preload("Product").Where("product_id = ?", productID).Find(&items).Error
}

func (r *ModelRepo) Search(query string) ([]frames.Model, error) {
	var items []frames.Model
	return items, r.DB.
		Preload("Product").
		Joins("JOIN product p ON p.id_product = model.product_id").
		Where("model.title_variant ILIKE ? OR p.title_product ILIKE ? OR model.upc ILIKE ? OR model.mfg_number ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Find(&items).Error
}

func (r *ModelRepo) Create(v *frames.Model) error {
	return r.DB.Create(v).Error
}

func (r *ModelRepo) Save(v *frames.Model) error {
	return r.DB.Save(v).Error
}

func (r *ModelRepo) Delete(id int64) error {
	return r.DB.Delete(&frames.Model{}, id).Error
}

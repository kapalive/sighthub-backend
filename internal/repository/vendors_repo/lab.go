package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type LabRepo struct{ DB *gorm.DB }

func NewLabRepo(db *gorm.DB) *LabRepo { return &LabRepo{DB: db} }

func (r *LabRepo) GetAll() ([]vendors.Lab, error) {
	var items []vendors.Lab
	return items, r.DB.Find(&items).Error
}

func (r *LabRepo) GetByID(id int) (*vendors.Lab, error) {
	var item vendors.Lab
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *LabRepo) GetInternal() ([]vendors.Lab, error) {
	var items []vendors.Lab
	return items, r.DB.Where("is_internal = true").Find(&items).Error
}

func (r *LabRepo) Create(item *vendors.Lab) error {
	return r.DB.Create(item).Error
}

func (r *LabRepo) Save(item *vendors.Lab) error {
	return r.DB.Save(item).Error
}

func (r *LabRepo) Delete(id int) error {
	return r.DB.Delete(&vendors.Lab{}, id).Error
}

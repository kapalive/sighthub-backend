package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type GroupPQRSRepo struct{ DB *gorm.DB }

func NewGroupPQRSRepo(db *gorm.DB) *GroupPQRSRepo { return &GroupPQRSRepo{DB: db} }

func (r *GroupPQRSRepo) GetAll() ([]general.GroupPQRS, error) {
	var items []general.GroupPQRS
	return items, r.DB.Order("title").Find(&items).Error
}

func (r *GroupPQRSRepo) GetByID(id int) (*general.GroupPQRS, error) {
	var v general.GroupPQRS
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *GroupPQRSRepo) Create(v *general.GroupPQRS) error { return r.DB.Create(v).Error }
func (r *GroupPQRSRepo) Save(v *general.GroupPQRS) error   { return r.DB.Save(v).Error }
func (r *GroupPQRSRepo) Delete(id int) error {
	return r.DB.Delete(&general.GroupPQRS{}, id).Error
}

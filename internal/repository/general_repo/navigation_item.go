package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type NavigationItemRepo struct{ DB *gorm.DB }

func NewNavigationItemRepo(db *gorm.DB) *NavigationItemRepo { return &NavigationItemRepo{DB: db} }

func (r *NavigationItemRepo) GetAll() ([]general.NavigationItem, error) {
	var items []general.NavigationItem
	return items, r.DB.Order("position ASC, id_navigation_item ASC").Find(&items).Error
}

func (r *NavigationItemRepo) GetByID(id int) (*general.NavigationItem, error) {
	var v general.NavigationItem
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *NavigationItemRepo) GetByGroupID(groupID int) ([]general.NavigationItem, error) {
	var items []general.NavigationItem
	return items, r.DB.Where("navigation_group_id = ?", groupID).Find(&items).Error
}

func (r *NavigationItemRepo) Create(v *general.NavigationItem) error { return r.DB.Create(v).Error }
func (r *NavigationItemRepo) Save(v *general.NavigationItem) error   { return r.DB.Save(v).Error }
func (r *NavigationItemRepo) Delete(id int) error {
	return r.DB.Delete(&general.NavigationItem{}, id).Error
}

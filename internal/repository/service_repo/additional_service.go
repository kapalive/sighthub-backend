package service_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/service"
)

type AdditionalServiceRepo struct{ DB *gorm.DB }

func NewAdditionalServiceRepo(db *gorm.DB) *AdditionalServiceRepo {
	return &AdditionalServiceRepo{DB: db}
}

func (r *AdditionalServiceRepo) GetByID(id int64) (*service.AdditionalService, error) {
	var item service.AdditionalService
	if err := r.DB.Preload("AddServiceType").First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *AdditionalServiceRepo) GetAll(visibleOnly bool) ([]service.AdditionalService, error) {
	var items []service.AdditionalService
	q := r.DB.Preload("AddServiceType")
	if visibleOnly {
		q = q.Where("visible = true")
	}
	return items, q.Order("sort1, sort2").Find(&items).Error
}

func (r *AdditionalServiceRepo) GetByTypeID(typeID int) ([]service.AdditionalService, error) {
	var items []service.AdditionalService
	return items, r.DB.Where("add_service_type_id = ?", typeID).Find(&items).Error
}

func (r *AdditionalServiceRepo) Search(query string) ([]service.AdditionalService, error) {
	var items []service.AdditionalService
	q := "%" + query + "%"
	return items, r.DB.Where("item_number ILIKE ? OR invoice_desc ILIKE ?", q, q).Find(&items).Error
}

func (r *AdditionalServiceRepo) Create(item *service.AdditionalService) error {
	return r.DB.Create(item).Error
}

func (r *AdditionalServiceRepo) Save(item *service.AdditionalService) error {
	return r.DB.Save(item).Error
}

func (r *AdditionalServiceRepo) Delete(id int64) error {
	return r.DB.Delete(&service.AdditionalService{}, id).Error
}

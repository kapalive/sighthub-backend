package service_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/service"
)

type ProfessionalServiceRepo struct{ DB *gorm.DB }

func NewProfessionalServiceRepo(db *gorm.DB) *ProfessionalServiceRepo {
	return &ProfessionalServiceRepo{DB: db}
}

func (r *ProfessionalServiceRepo) GetByID(id int64) (*service.ProfessionalService, error) {
	var item service.ProfessionalService
	if err := r.DB.Preload("Scope").Preload("Type").First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ProfessionalServiceRepo) GetAll(visibleOnly bool) ([]service.ProfessionalService, error) {
	var items []service.ProfessionalService
	q := r.DB.Preload("Scope").Preload("Type")
	if visibleOnly {
		q = q.Where("visible = true")
	}
	return items, q.Order("sort1, sort2").Find(&items).Error
}

func (r *ProfessionalServiceRepo) GetByTypeID(typeID int) ([]service.ProfessionalService, error) {
	var items []service.ProfessionalService
	return items, r.DB.Where("professional_service_type_id = ?", typeID).Find(&items).Error
}

func (r *ProfessionalServiceRepo) Search(query string) ([]service.ProfessionalService, error) {
	var items []service.ProfessionalService
	q := "%" + query + "%"
	return items, r.DB.Where("item_number ILIKE ? OR invoice_desc ILIKE ? OR cpt_hcpcs_code ILIKE ?", q, q, q).
		Find(&items).Error
}

func (r *ProfessionalServiceRepo) Create(item *service.ProfessionalService) error {
	return r.DB.Create(item).Error
}

func (r *ProfessionalServiceRepo) Save(item *service.ProfessionalService) error {
	return r.DB.Save(item).Error
}

func (r *ProfessionalServiceRepo) Delete(id int64) error {
	return r.DB.Delete(&service.ProfessionalService{}, id).Error
}

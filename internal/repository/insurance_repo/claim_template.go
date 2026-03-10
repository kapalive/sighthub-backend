package insurance_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/insurance"
)

type ClaimTemplateRepo struct{ DB *gorm.DB }

func NewClaimTemplateRepo(db *gorm.DB) *ClaimTemplateRepo { return &ClaimTemplateRepo{DB: db} }

func (r *ClaimTemplateRepo) GetByID(id int) (*insurance.ClaimTemplate, error) {
	var v insurance.ClaimTemplate
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *ClaimTemplateRepo) GetByLocationID(locationID int) ([]insurance.ClaimTemplate, error) {
	var items []insurance.ClaimTemplate
	return items, r.DB.Where("location_id = ?", locationID).Order("name").Find(&items).Error
}

func (r *ClaimTemplateRepo) GetByDoctorID(doctorID int) ([]insurance.ClaimTemplate, error) {
	var items []insurance.ClaimTemplate
	return items, r.DB.Where("doctor_id = ?", doctorID).Order("name").Find(&items).Error
}

func (r *ClaimTemplateRepo) Create(v *insurance.ClaimTemplate) error { return r.DB.Create(v).Error }
func (r *ClaimTemplateRepo) Save(v *insurance.ClaimTemplate) error   { return r.DB.Save(v).Error }
func (r *ClaimTemplateRepo) Delete(id int) error {
	return r.DB.Delete(&insurance.ClaimTemplate{}, id).Error
}

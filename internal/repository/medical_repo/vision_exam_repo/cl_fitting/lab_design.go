package cl_fitting

import (
	"gorm.io/gorm"
	cl "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
)

type LabDesignRepo struct{ DB *gorm.DB }
func NewLabDesignRepo(db *gorm.DB) *LabDesignRepo { return &LabDesignRepo{DB: db} }
func (r *LabDesignRepo) GetByID(id int64) (*cl.LabDesign, error) {
	var v cl.LabDesign
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *LabDesignRepo) Create() (*cl.LabDesign, error) {
	v := cl.LabDesign{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *LabDesignRepo) Save(v *cl.LabDesign) error { return r.DB.Save(v).Error }

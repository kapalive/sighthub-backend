package cl_fitting

import (
	"gorm.io/gorm"
	cl "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
)

type GasPermeableRepo struct{ DB *gorm.DB }
func NewGasPermeableRepo(db *gorm.DB) *GasPermeableRepo { return &GasPermeableRepo{DB: db} }
func (r *GasPermeableRepo) GetByID(id int64) (*cl.GasPermeable, error) {
	var v cl.GasPermeable
	if err := r.DB.Preload("LabDesign").Preload("DrDesign").First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}
func (r *GasPermeableRepo) Create() (*cl.GasPermeable, error) {
	v := cl.GasPermeable{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *GasPermeableRepo) Save(v *cl.GasPermeable) error { return r.DB.Save(v).Error }

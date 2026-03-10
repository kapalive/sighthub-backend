package cl_fitting

import (
	"gorm.io/gorm"
	cl "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
)

type DrDesignRepo struct{ DB *gorm.DB }
func NewDrDesignRepo(db *gorm.DB) *DrDesignRepo { return &DrDesignRepo{DB: db} }
func (r *DrDesignRepo) GetByID(id int64) (*cl.DrDesign, error) {
	var v cl.DrDesign
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *DrDesignRepo) Create() (*cl.DrDesign, error) {
	v := cl.DrDesign{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *DrDesignRepo) Save(v *cl.DrDesign) error { return r.DB.Save(v).Error }

package cl_fitting

import (
	"gorm.io/gorm"
	cl "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
)

type Fitting1Repo struct{ DB *gorm.DB }
func NewFitting1Repo(db *gorm.DB) *Fitting1Repo { return &Fitting1Repo{DB: db} }
func (r *Fitting1Repo) GetByID(id int64) (*cl.Fitting1, error) {
	var v cl.Fitting1
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *Fitting1Repo) Create() (*cl.Fitting1, error) {
	v := cl.Fitting1{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *Fitting1Repo) Save(v *cl.Fitting1) error { return r.DB.Save(v).Error }

package cl_fitting

import (
	"gorm.io/gorm"
	cl "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
)

type Fitting3Repo struct{ DB *gorm.DB }
func NewFitting3Repo(db *gorm.DB) *Fitting3Repo { return &Fitting3Repo{DB: db} }
func (r *Fitting3Repo) GetByID(id int64) (*cl.Fitting3, error) {
	var v cl.Fitting3
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *Fitting3Repo) Create() (*cl.Fitting3, error) {
	v := cl.Fitting3{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *Fitting3Repo) Save(v *cl.Fitting3) error { return r.DB.Save(v).Error }

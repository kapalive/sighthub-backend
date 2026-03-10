package cl_fitting

import (
	"gorm.io/gorm"
	cl "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
)

type Fitting2Repo struct{ DB *gorm.DB }
func NewFitting2Repo(db *gorm.DB) *Fitting2Repo { return &Fitting2Repo{DB: db} }
func (r *Fitting2Repo) GetByID(id int64) (*cl.Fitting2, error) {
	var v cl.Fitting2
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *Fitting2Repo) Create() (*cl.Fitting2, error) {
	v := cl.Fitting2{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *Fitting2Repo) Save(v *cl.Fitting2) error { return r.DB.Save(v).Error }

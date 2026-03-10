package cl_fitting

import (
	"gorm.io/gorm"
	cl "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
)

type SecondTrialRepo struct{ DB *gorm.DB }
func NewSecondTrialRepo(db *gorm.DB) *SecondTrialRepo { return &SecondTrialRepo{DB: db} }
func (r *SecondTrialRepo) GetByID(id int64) (*cl.SecondTrial, error) {
	var v cl.SecondTrial
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *SecondTrialRepo) Create() (*cl.SecondTrial, error) {
	v := cl.SecondTrial{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *SecondTrialRepo) Save(v *cl.SecondTrial) error { return r.DB.Save(v).Error }

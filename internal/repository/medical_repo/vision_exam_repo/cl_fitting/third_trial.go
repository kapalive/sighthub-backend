package cl_fitting

import (
	"gorm.io/gorm"
	cl "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
)

type ThirdTrialRepo struct{ DB *gorm.DB }
func NewThirdTrialRepo(db *gorm.DB) *ThirdTrialRepo { return &ThirdTrialRepo{DB: db} }
func (r *ThirdTrialRepo) GetByID(id int64) (*cl.ThirdTrial, error) {
	var v cl.ThirdTrial
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *ThirdTrialRepo) Create() (*cl.ThirdTrial, error) {
	v := cl.ThirdTrial{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *ThirdTrialRepo) Save(v *cl.ThirdTrial) error { return r.DB.Save(v).Error }

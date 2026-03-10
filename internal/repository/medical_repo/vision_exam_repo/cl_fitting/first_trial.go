package cl_fitting

import (
	"gorm.io/gorm"
	cl "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
)

type FirstTrialRepo struct{ DB *gorm.DB }
func NewFirstTrialRepo(db *gorm.DB) *FirstTrialRepo { return &FirstTrialRepo{DB: db} }
func (r *FirstTrialRepo) GetByID(id int64) (*cl.FirstTrial, error) {
	var v cl.FirstTrial
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *FirstTrialRepo) Create() (*cl.FirstTrial, error) {
	v := cl.FirstTrial{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *FirstTrialRepo) Save(v *cl.FirstTrial) error { return r.DB.Save(v).Error }

package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type BloodPressureRepo struct{ DB *gorm.DB }
func NewBloodPressureRepo(db *gorm.DB) *BloodPressureRepo { return &BloodPressureRepo{DB: db} }

func (r *BloodPressureRepo) GetByID(id int64) (*p.BloodPressure, error) {
	var v p.BloodPressure
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *BloodPressureRepo) Create() (*p.BloodPressure, error) {
	v := p.BloodPressure{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *BloodPressureRepo) Save(v *p.BloodPressure) error { return r.DB.Save(v).Error }

package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type ColorVisionRepo struct{ DB *gorm.DB }
func NewColorVisionRepo(db *gorm.DB) *ColorVisionRepo { return &ColorVisionRepo{DB: db} }

func (r *ColorVisionRepo) GetByID(id int64) (*p.ColorVision, error) {
	var v p.ColorVision
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *ColorVisionRepo) Create() (*p.ColorVision, error) {
	v := p.ColorVision{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *ColorVisionRepo) Save(v *p.ColorVision) error { return r.DB.Save(v).Error }

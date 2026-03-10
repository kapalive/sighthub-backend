package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type UnaidedVANearRepo struct{ DB *gorm.DB }
func NewUnaidedVANearRepo(db *gorm.DB) *UnaidedVANearRepo { return &UnaidedVANearRepo{DB: db} }

func (r *UnaidedVANearRepo) GetByID(id int64) (*p.UnaidedVANear, error) {
	var v p.UnaidedVANear
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *UnaidedVANearRepo) Create() (*p.UnaidedVANear, error) {
	v := p.UnaidedVANear{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *UnaidedVANearRepo) Save(v *p.UnaidedVANear) error { return r.DB.Save(v).Error }

package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type AidedVANearRepo struct{ DB *gorm.DB }
func NewAidedVANearRepo(db *gorm.DB) *AidedVANearRepo { return &AidedVANearRepo{DB: db} }

func (r *AidedVANearRepo) GetByID(id int64) (*p.AidedVANear, error) {
	var v p.AidedVANear
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AidedVANearRepo) Create() (*p.AidedVANear, error) {
	v := p.AidedVANear{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AidedVANearRepo) Save(v *p.AidedVANear) error { return r.DB.Save(v).Error }

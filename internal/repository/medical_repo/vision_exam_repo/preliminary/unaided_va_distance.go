package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type UnaidedVADistanceRepo struct{ DB *gorm.DB }
func NewUnaidedVADistanceRepo(db *gorm.DB) *UnaidedVADistanceRepo { return &UnaidedVADistanceRepo{DB: db} }

func (r *UnaidedVADistanceRepo) GetByID(id int64) (*p.UnaidedVADistance, error) {
	var v p.UnaidedVADistance
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *UnaidedVADistanceRepo) Create() (*p.UnaidedVADistance, error) {
	v := p.UnaidedVADistance{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *UnaidedVADistanceRepo) Save(v *p.UnaidedVADistance) error { return r.DB.Save(v).Error }

package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type UnaidedPHDistanceRepo struct{ DB *gorm.DB }
func NewUnaidedPHDistanceRepo(db *gorm.DB) *UnaidedPHDistanceRepo { return &UnaidedPHDistanceRepo{DB: db} }

func (r *UnaidedPHDistanceRepo) GetByID(id int64) (*p.UnaidedPHDistance, error) {
	var v p.UnaidedPHDistance
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *UnaidedPHDistanceRepo) Create() (*p.UnaidedPHDistance, error) {
	v := p.UnaidedPHDistance{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *UnaidedPHDistanceRepo) Save(v *p.UnaidedPHDistance) error { return r.DB.Save(v).Error }

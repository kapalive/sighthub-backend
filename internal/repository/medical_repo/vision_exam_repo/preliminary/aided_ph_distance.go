package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type AidedPHDistanceRepo struct{ DB *gorm.DB }
func NewAidedPHDistanceRepo(db *gorm.DB) *AidedPHDistanceRepo { return &AidedPHDistanceRepo{DB: db} }

func (r *AidedPHDistanceRepo) GetByID(id int64) (*p.AidedPHDistance, error) {
	var v p.AidedPHDistance
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AidedPHDistanceRepo) Create() (*p.AidedPHDistance, error) {
	v := p.AidedPHDistance{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AidedPHDistanceRepo) Save(v *p.AidedPHDistance) error { return r.DB.Save(v).Error }

package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type AidedVADistanceRepo struct{ DB *gorm.DB }
func NewAidedVADistanceRepo(db *gorm.DB) *AidedVADistanceRepo { return &AidedVADistanceRepo{DB: db} }

func (r *AidedVADistanceRepo) GetByID(id int64) (*p.AidedVADistance, error) {
	var v p.AidedVADistance
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AidedVADistanceRepo) Create() (*p.AidedVADistance, error) {
	v := p.AidedVADistance{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AidedVADistanceRepo) Save(v *p.AidedVADistance) error { return r.DB.Save(v).Error }

package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type AmslerGridRepo struct{ DB *gorm.DB }
func NewAmslerGridRepo(db *gorm.DB) *AmslerGridRepo { return &AmslerGridRepo{DB: db} }

func (r *AmslerGridRepo) GetByID(id int64) (*p.AmslerGrid, error) {
	var v p.AmslerGrid
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AmslerGridRepo) Create() (*p.AmslerGrid, error) {
	v := p.AmslerGrid{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AmslerGridRepo) Save(v *p.AmslerGrid) error { return r.DB.Save(v).Error }

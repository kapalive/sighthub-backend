package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type MotilityRepo struct{ DB *gorm.DB }
func NewMotilityRepo(db *gorm.DB) *MotilityRepo { return &MotilityRepo{DB: db} }

func (r *MotilityRepo) GetByID(id int64) (*p.Motility, error) {
	var v p.Motility
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *MotilityRepo) Create() (*p.Motility, error) {
	v := p.Motility{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *MotilityRepo) Save(v *p.Motility) error { return r.DB.Save(v).Error }

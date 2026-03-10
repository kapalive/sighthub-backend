package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type AccommodationRepo struct{ DB *gorm.DB }
func NewAccommodationRepo(db *gorm.DB) *AccommodationRepo { return &AccommodationRepo{DB: db} }

func (r *AccommodationRepo) GetByID(id int64) (*p.Accommodation, error) {
	var v p.Accommodation
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AccommodationRepo) Create() (*p.Accommodation, error) {
	v := p.Accommodation{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AccommodationRepo) Save(v *p.Accommodation) error { return r.DB.Save(v).Error }

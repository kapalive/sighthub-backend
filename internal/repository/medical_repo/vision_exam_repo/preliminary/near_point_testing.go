package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type NearPointTestingRepo struct{ DB *gorm.DB }
func NewNearPointTestingRepo(db *gorm.DB) *NearPointTestingRepo { return &NearPointTestingRepo{DB: db} }

func (r *NearPointTestingRepo) GetByID(id int64) (*p.NearPointTesting, error) {
	var v p.NearPointTesting
	if err := r.DB.
		Preload("DistPhoria").
		Preload("NearPhoria").
		Preload("DistVergence").
		Preload("NearVergence").
		Preload("Accommodation").
		First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}
func (r *NearPointTestingRepo) Create() (*p.NearPointTesting, error) {
	v := p.NearPointTesting{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *NearPointTestingRepo) Save(v *p.NearPointTesting) error { return r.DB.Save(v).Error }

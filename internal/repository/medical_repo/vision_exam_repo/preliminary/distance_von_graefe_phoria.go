package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type DistanceVonGraefePhoriaRepo struct{ DB *gorm.DB }
func NewDistanceVonGraefePhoriaRepo(db *gorm.DB) *DistanceVonGraefePhoriaRepo { return &DistanceVonGraefePhoriaRepo{DB: db} }

func (r *DistanceVonGraefePhoriaRepo) GetByID(id int64) (*p.DistanceVonGraefePhoria, error) {
	var v p.DistanceVonGraefePhoria
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *DistanceVonGraefePhoriaRepo) Create() (*p.DistanceVonGraefePhoria, error) {
	v := p.DistanceVonGraefePhoria{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *DistanceVonGraefePhoriaRepo) Save(v *p.DistanceVonGraefePhoria) error { return r.DB.Save(v).Error }

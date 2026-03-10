package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type NearVonGraefePhoriaRepo struct{ DB *gorm.DB }
func NewNearVonGraefePhoriaRepo(db *gorm.DB) *NearVonGraefePhoriaRepo { return &NearVonGraefePhoriaRepo{DB: db} }

func (r *NearVonGraefePhoriaRepo) GetByID(id int64) (*p.NearVonGraefePhoria, error) {
	var v p.NearVonGraefePhoria
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *NearVonGraefePhoriaRepo) Create() (*p.NearVonGraefePhoria, error) {
	v := p.NearVonGraefePhoria{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *NearVonGraefePhoriaRepo) Save(v *p.NearVonGraefePhoria) error { return r.DB.Save(v).Error }

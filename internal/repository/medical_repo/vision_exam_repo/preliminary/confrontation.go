package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type ConfrontationRepo struct{ DB *gorm.DB }
func NewConfrontationRepo(db *gorm.DB) *ConfrontationRepo { return &ConfrontationRepo{DB: db} }

func (r *ConfrontationRepo) GetByID(id int64) (*p.Confrontation, error) {
	var v p.Confrontation
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *ConfrontationRepo) Create() (*p.Confrontation, error) {
	v := p.Confrontation{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *ConfrontationRepo) Save(v *p.Confrontation) error { return r.DB.Save(v).Error }

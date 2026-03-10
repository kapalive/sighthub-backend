package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type EntranceGlassesRepo struct{ DB *gorm.DB }
func NewEntranceGlassesRepo(db *gorm.DB) *EntranceGlassesRepo { return &EntranceGlassesRepo{DB: db} }

func (r *EntranceGlassesRepo) GetByID(id int64) (*p.EntranceGlasses, error) {
	var v p.EntranceGlasses
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *EntranceGlassesRepo) Create() (*p.EntranceGlasses, error) {
	v := p.EntranceGlasses{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *EntranceGlassesRepo) Save(v *p.EntranceGlasses) error { return r.DB.Save(v).Error }

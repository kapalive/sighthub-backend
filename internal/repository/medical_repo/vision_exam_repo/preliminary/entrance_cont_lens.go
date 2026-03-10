package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type EntranceContLensRepo struct{ DB *gorm.DB }
func NewEntranceContLensRepo(db *gorm.DB) *EntranceContLensRepo { return &EntranceContLensRepo{DB: db} }

func (r *EntranceContLensRepo) GetByID(id int64) (*p.EntranceContLens, error) {
	var v p.EntranceContLens
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *EntranceContLensRepo) Create() (*p.EntranceContLens, error) {
	v := p.EntranceContLens{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *EntranceContLensRepo) Save(v *p.EntranceContLens) error { return r.DB.Save(v).Error }

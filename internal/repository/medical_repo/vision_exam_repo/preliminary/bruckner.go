package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type BrucknerRepo struct{ DB *gorm.DB }
func NewBrucknerRepo(db *gorm.DB) *BrucknerRepo { return &BrucknerRepo{DB: db} }

func (r *BrucknerRepo) GetByID(id int64) (*p.Bruckner, error) {
	var v p.Bruckner
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *BrucknerRepo) Create() (*p.Bruckner, error) {
	v := p.Bruckner{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *BrucknerRepo) Save(v *p.Bruckner) error { return r.DB.Save(v).Error }

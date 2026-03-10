package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type AutoKeratometerPreliminaryRepo struct{ DB *gorm.DB }
func NewAutoKeratometerPreliminaryRepo(db *gorm.DB) *AutoKeratometerPreliminaryRepo { return &AutoKeratometerPreliminaryRepo{DB: db} }

func (r *AutoKeratometerPreliminaryRepo) GetByID(id int64) (*p.AutoKeratometerPreliminary, error) {
	var v p.AutoKeratometerPreliminary
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AutoKeratometerPreliminaryRepo) Create() (*p.AutoKeratometerPreliminary, error) {
	v := p.AutoKeratometerPreliminary{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AutoKeratometerPreliminaryRepo) Save(v *p.AutoKeratometerPreliminary) error { return r.DB.Save(v).Error }

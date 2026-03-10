package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type AutorefractorPreliminaryRepo struct{ DB *gorm.DB }
func NewAutorefractorPreliminaryRepo(db *gorm.DB) *AutorefractorPreliminaryRepo { return &AutorefractorPreliminaryRepo{DB: db} }

func (r *AutorefractorPreliminaryRepo) GetByID(id int64) (*p.AutorefractorPreliminary, error) {
	var v p.AutorefractorPreliminary
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AutorefractorPreliminaryRepo) Create() (*p.AutorefractorPreliminary, error) {
	v := p.AutorefractorPreliminary{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AutorefractorPreliminaryRepo) Save(v *p.AutorefractorPreliminary) error { return r.DB.Save(v).Error }

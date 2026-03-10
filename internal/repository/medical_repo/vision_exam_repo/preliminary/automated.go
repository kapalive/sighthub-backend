package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type AutomatedRepo struct{ DB *gorm.DB }
func NewAutomatedRepo(db *gorm.DB) *AutomatedRepo { return &AutomatedRepo{DB: db} }

func (r *AutomatedRepo) GetByID(id int64) (*p.Automated, error) {
	var v p.Automated
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AutomatedRepo) Create() (*p.Automated, error) {
	v := p.Automated{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *AutomatedRepo) Save(v *p.Automated) error { return r.DB.Save(v).Error }

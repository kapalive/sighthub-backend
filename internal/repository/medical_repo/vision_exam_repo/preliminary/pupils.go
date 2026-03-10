package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type PupilsRepo struct{ DB *gorm.DB }
func NewPupilsRepo(db *gorm.DB) *PupilsRepo { return &PupilsRepo{DB: db} }

func (r *PupilsRepo) GetByID(id int64) (*p.Pupils, error) {
	var v p.Pupils
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *PupilsRepo) Create() (*p.Pupils, error) {
	v := p.Pupils{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *PupilsRepo) Save(v *p.Pupils) error { return r.DB.Save(v).Error }

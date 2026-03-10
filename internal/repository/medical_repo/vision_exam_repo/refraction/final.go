package refraction

import (
	"gorm.io/gorm"
	r "sighthub-backend/internal/models/medical/vision_exam/refraction"
)

type RefractionFinalRepo struct{ DB *gorm.DB }
func NewRefractionFinalRepo(db *gorm.DB) *RefractionFinalRepo { return &RefractionFinalRepo{DB: db} }
func (repo *RefractionFinalRepo) GetByID(id int64) (*r.RefractionFinal, error) {
	var v r.RefractionFinal
	if err := repo.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *RefractionFinalRepo) Create() (*r.RefractionFinal, error) {
	v := r.RefractionFinal{}
	if err := repo.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *RefractionFinalRepo) Save(v *r.RefractionFinal) error { return repo.DB.Save(v).Error }

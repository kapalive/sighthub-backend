package refraction

import (
	"gorm.io/gorm"
	r "sighthub-backend/internal/models/medical/vision_exam/refraction"
)

type CycloRepo struct{ DB *gorm.DB }
func NewCycloRepo(db *gorm.DB) *CycloRepo { return &CycloRepo{DB: db} }
func (repo *CycloRepo) GetByID(id int64) (*r.Cyclo, error) {
	var v r.Cyclo
	if err := repo.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *CycloRepo) Create() (*r.Cyclo, error) {
	v := r.Cyclo{}
	if err := repo.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *CycloRepo) Save(v *r.Cyclo) error { return repo.DB.Save(v).Error }

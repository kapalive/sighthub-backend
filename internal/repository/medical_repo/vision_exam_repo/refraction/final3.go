package refraction

import (
	"gorm.io/gorm"
	r "sighthub-backend/internal/models/medical/vision_exam/refraction"
)

type Final3Repo struct{ DB *gorm.DB }
func NewFinal3Repo(db *gorm.DB) *Final3Repo { return &Final3Repo{DB: db} }
func (repo *Final3Repo) GetByID(id int64) (*r.Final3, error) {
	var v r.Final3
	if err := repo.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *Final3Repo) Create() (*r.Final3, error) {
	v := r.Final3{}
	if err := repo.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *Final3Repo) Save(v *r.Final3) error { return repo.DB.Save(v).Error }

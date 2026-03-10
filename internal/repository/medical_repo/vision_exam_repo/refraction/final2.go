package refraction

import (
	"gorm.io/gorm"
	r "sighthub-backend/internal/models/medical/vision_exam/refraction"
)

type Final2Repo struct{ DB *gorm.DB }
func NewFinal2Repo(db *gorm.DB) *Final2Repo { return &Final2Repo{DB: db} }
func (repo *Final2Repo) GetByID(id int64) (*r.Final2, error) {
	var v r.Final2
	if err := repo.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *Final2Repo) Create() (*r.Final2, error) {
	v := r.Final2{}
	if err := repo.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *Final2Repo) Save(v *r.Final2) error { return repo.DB.Save(v).Error }

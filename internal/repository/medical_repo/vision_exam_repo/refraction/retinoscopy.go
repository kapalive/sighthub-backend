package refraction

import (
	"gorm.io/gorm"
	r "sighthub-backend/internal/models/medical/vision_exam/refraction"
)

type RetinoscopyRepo struct{ DB *gorm.DB }
func NewRetinoscopyRepo(db *gorm.DB) *RetinoscopyRepo { return &RetinoscopyRepo{DB: db} }
func (repo *RetinoscopyRepo) GetByID(id int64) (*r.Retinoscopy, error) {
	var v r.Retinoscopy
	if err := repo.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *RetinoscopyRepo) Create() (*r.Retinoscopy, error) {
	v := r.Retinoscopy{}
	if err := repo.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *RetinoscopyRepo) Save(v *r.Retinoscopy) error { return repo.DB.Save(v).Error }

package refraction

import (
	"gorm.io/gorm"
	r "sighthub-backend/internal/models/medical/vision_exam/refraction"
)

type ManifestRepo struct{ DB *gorm.DB }
func NewManifestRepo(db *gorm.DB) *ManifestRepo { return &ManifestRepo{DB: db} }
func (repo *ManifestRepo) GetByID(id int64) (*r.Manifest, error) {
	var v r.Manifest
	if err := repo.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *ManifestRepo) Create() (*r.Manifest, error) {
	v := r.Manifest{}
	if err := repo.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (repo *ManifestRepo) Save(v *r.Manifest) error { return repo.DB.Save(v).Error }

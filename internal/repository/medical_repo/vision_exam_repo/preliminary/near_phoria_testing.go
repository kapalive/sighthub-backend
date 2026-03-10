package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type NearPhoriaTestRepo struct{ DB *gorm.DB }
func NewNearPhoriaTestRepo(db *gorm.DB) *NearPhoriaTestRepo { return &NearPhoriaTestRepo{DB: db} }

func (r *NearPhoriaTestRepo) GetByID(id int64) (*p.NearPhoriaTest, error) {
	var v p.NearPhoriaTest
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *NearPhoriaTestRepo) Create() (*p.NearPhoriaTest, error) {
	v := p.NearPhoriaTest{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *NearPhoriaTestRepo) Save(v *p.NearPhoriaTest) error { return r.DB.Save(v).Error }

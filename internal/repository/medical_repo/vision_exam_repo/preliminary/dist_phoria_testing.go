package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type DistPhoriaTestRepo struct{ DB *gorm.DB }
func NewDistPhoriaTestRepo(db *gorm.DB) *DistPhoriaTestRepo { return &DistPhoriaTestRepo{DB: db} }

func (r *DistPhoriaTestRepo) GetByID(id int64) (*p.DistPhoriaTest, error) {
	var v p.DistPhoriaTest
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *DistPhoriaTestRepo) Create() (*p.DistPhoriaTest, error) {
	v := p.DistPhoriaTest{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *DistPhoriaTestRepo) Save(v *p.DistPhoriaTest) error { return r.DB.Save(v).Error }

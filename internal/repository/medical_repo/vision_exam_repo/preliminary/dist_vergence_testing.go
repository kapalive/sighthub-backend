package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type DistVergenceTestRepo struct{ DB *gorm.DB }
func NewDistVergenceTestRepo(db *gorm.DB) *DistVergenceTestRepo { return &DistVergenceTestRepo{DB: db} }

func (r *DistVergenceTestRepo) GetByID(id int64) (*p.DistVergenceTest, error) {
	var v p.DistVergenceTest
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *DistVergenceTestRepo) Create() (*p.DistVergenceTest, error) {
	v := p.DistVergenceTest{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *DistVergenceTestRepo) Save(v *p.DistVergenceTest) error { return r.DB.Save(v).Error }

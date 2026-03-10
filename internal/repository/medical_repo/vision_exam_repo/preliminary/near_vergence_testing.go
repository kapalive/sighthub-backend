package preliminary

import (
	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type NearVergenceTestRepo struct{ DB *gorm.DB }
func NewNearVergenceTestRepo(db *gorm.DB) *NearVergenceTestRepo { return &NearVergenceTestRepo{DB: db} }

func (r *NearVergenceTestRepo) GetByID(id int64) (*p.NearVergenceTest, error) {
	var v p.NearVergenceTest
	if err := r.DB.First(&v, id).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *NearVergenceTestRepo) Create() (*p.NearVergenceTest, error) {
	v := p.NearVergenceTest{}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}
func (r *NearVergenceTestRepo) Save(v *p.NearVergenceTest) error { return r.DB.Save(v).Error }

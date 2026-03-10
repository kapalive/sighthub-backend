// internal/repository/medical_repo/vision_exam_repo/history/social_history.go
package history

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/medical/vision_exam/history"
)

type SocialHistoryRepo struct{ DB *gorm.DB }

func NewSocialHistoryRepo(db *gorm.DB) *SocialHistoryRepo {
	return &SocialHistoryRepo{DB: db}
}

func (r *SocialHistoryRepo) GetByID(id int64) (*history.SocialHistory, error) {
	var v history.SocialHistory
	if err := r.DB.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *SocialHistoryRepo) Create() (*history.SocialHistory, error) {
	v := history.SocialHistory{}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *SocialHistoryRepo) Save(v *history.SocialHistory) error {
	return r.DB.Save(v).Error
}

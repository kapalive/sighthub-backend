package referral

import (
	"errors"

	"gorm.io/gorm"
	ref "sighthub-backend/internal/models/medical/vision_exam/referral"
)

type ReferralLetterRepo struct{ DB *gorm.DB }

func NewReferralLetterRepo(db *gorm.DB) *ReferralLetterRepo {
	return &ReferralLetterRepo{DB: db}
}

func (r *ReferralLetterRepo) GetByEyeExamID(eyeExamID int64) (*ref.ReferralLetter, error) {
	var v ref.ReferralLetter
	if err := r.DB.
		Preload("ToDoctor").
		Preload("CcDoctor").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ReferralLetterRepo) GetByID(id int64) (*ref.ReferralLetter, error) {
	var v ref.ReferralLetter
	if err := r.DB.
		Preload("ToDoctor").
		Preload("CcDoctor").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ReferralLetterRepo) Create(eyeExamID int64) (*ref.ReferralLetter, error) {
	v := ref.ReferralLetter{EyeExamID: eyeExamID}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ReferralLetterRepo) Save(v *ref.ReferralLetter) error {
	return r.DB.Save(v).Error
}

package posterior

import (
	"errors"

	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/posterior"
)

type PosteriorEyeRepo struct{ DB *gorm.DB }

func NewPosteriorEyeRepo(db *gorm.DB) *PosteriorEyeRepo {
	return &PosteriorEyeRepo{DB: db}
}

func (r *PosteriorEyeRepo) GetByEyeExamID(eyeExamID int64) (*p.PosteriorEye, error) {
	var v p.PosteriorEye
	if err := r.DB.
		Preload("Findings").
		Preload("CupDiscRatio").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *PosteriorEyeRepo) GetByID(id int64) (*p.PosteriorEye, error) {
	var v p.PosteriorEye
	if err := r.DB.
		Preload("Findings").
		Preload("CupDiscRatio").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *PosteriorEyeRepo) Create(eyeExamID, findingsID, cupDiscID int64) (*p.PosteriorEye, error) {
	v := p.PosteriorEye{
		EyeExamID:               eyeExamID,
		FindingsPosteriorID:     findingsID,
		CupDiscRatioPosteriorID: cupDiscID,
	}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *PosteriorEyeRepo) Save(v *p.PosteriorEye) error {
	return r.DB.Save(v).Error
}

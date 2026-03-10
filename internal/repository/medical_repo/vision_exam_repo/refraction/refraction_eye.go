package refraction

import (
	"errors"

	"gorm.io/gorm"
	r "sighthub-backend/internal/models/medical/vision_exam/refraction"
)

type RefractionEyeRepo struct{ DB *gorm.DB }
func NewRefractionEyeRepo(db *gorm.DB) *RefractionEyeRepo { return &RefractionEyeRepo{DB: db} }

func (repo *RefractionEyeRepo) GetByEyeExamID(eyeExamID int64) (*r.RefractionEye, error) {
	var v r.RefractionEye
	if err := repo.DB.
		Preload("Retinoscopy").
		Preload("Cyclo").
		Preload("Manifest").
		Preload("Final").
		Preload("Final2").
		Preload("Final3").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (repo *RefractionEyeRepo) Create(eyeExamID, retinoscopyID, cycloID, manifestID, finalID int64) (*r.RefractionEye, error) {
	v := r.RefractionEye{
		EyeExamID:     eyeExamID,
		RetinoscopyID: retinoscopyID,
		CycloID:       cycloID,
		ManifestID:    manifestID,
		FinalID:       finalID,
	}
	if err := repo.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}

func (repo *RefractionEyeRepo) Save(v *r.RefractionEye) error { return repo.DB.Save(v).Error }

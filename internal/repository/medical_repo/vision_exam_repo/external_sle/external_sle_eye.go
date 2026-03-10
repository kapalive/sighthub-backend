package external_sle

import (
	"errors"

	"gorm.io/gorm"
	e "sighthub-backend/internal/models/medical/vision_exam/external_sle"
)

type ExternalSleEyeRepo struct{ DB *gorm.DB }

func NewExternalSleEyeRepo(db *gorm.DB) *ExternalSleEyeRepo {
	return &ExternalSleEyeRepo{DB: db}
}

func (r *ExternalSleEyeRepo) GetByEyeExamID(eyeExamID int64) (*e.ExternalSleEye, error) {
	var v e.ExternalSleEye
	if err := r.DB.
		Preload("Findings").
		Preload("Gonioscopy").
		Preload("Pach").
		Preload("VisualFields").
		Preload("TonometryEyes").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ExternalSleEyeRepo) GetByID(id int64) (*e.ExternalSleEye, error) {
	var v e.ExternalSleEye
	if err := r.DB.
		Preload("Findings").
		Preload("Gonioscopy").
		Preload("Pach").
		Preload("VisualFields").
		Preload("TonometryEyes").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

type CreateExternalSleEyeInput struct {
	EyeExamID               int64
	FindingsExternalSleID   int64
	GonioscopyExternalSleID int64
	PachExternalSleID       int64
}

func (r *ExternalSleEyeRepo) Create(in CreateExternalSleEyeInput) (*e.ExternalSleEye, error) {
	v := e.ExternalSleEye{
		EyeExamID:               in.EyeExamID,
		FindingsExternalSleID:   in.FindingsExternalSleID,
		GonioscopyExternalSleID: in.GonioscopyExternalSleID,
		PachExternalSleID:       in.PachExternalSleID,
		OdAngleEstimation:       "n/a",
		OsAngleEstimation:       "n/a",
	}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ExternalSleEyeRepo) Save(v *e.ExternalSleEye) error {
	return r.DB.Save(v).Error
}

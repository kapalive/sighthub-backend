package cl_fitting

import (
	"errors"

	"gorm.io/gorm"
	cl "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
)

type ClFittingRepo struct{ DB *gorm.DB }
func NewClFittingRepo(db *gorm.DB) *ClFittingRepo { return &ClFittingRepo{DB: db} }

func (r *ClFittingRepo) GetByEyeExamID(eyeExamID int64) (*cl.ClFitting, error) {
	var v cl.ClFitting
	if err := r.DB.
		Preload("Fitting1").
		Preload("Fitting2").
		Preload("Fitting3").
		Preload("FirstTrial").
		Preload("SecondTrial").
		Preload("ThirdTrial").
		Preload("GasPermeable").
		Preload("GasPermeable.LabDesign").
		Preload("GasPermeable.DrDesign").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *ClFittingRepo) Create(eyeExamID, fitting1ID, firstTrialID int64) (*cl.ClFitting, error) {
	v := cl.ClFitting{
		EyeExamID:    eyeExamID,
		Fitting1ID:   fitting1ID,
		FirstTrialID: firstTrialID,
	}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}

func (r *ClFittingRepo) Save(v *cl.ClFitting) error { return r.DB.Save(v).Error }

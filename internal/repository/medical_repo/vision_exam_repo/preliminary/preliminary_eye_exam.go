package preliminary

import (
	"errors"

	"gorm.io/gorm"
	p "sighthub-backend/internal/models/medical/vision_exam/preliminary"
)

type PreliminaryEyeExamRepo struct{ DB *gorm.DB }
func NewPreliminaryEyeExamRepo(db *gorm.DB) *PreliminaryEyeExamRepo {
	return &PreliminaryEyeExamRepo{DB: db}
}

func (r *PreliminaryEyeExamRepo) GetByEyeExamID(eyeExamID int64) (*p.PreliminaryEyeExam, error) {
	var v p.PreliminaryEyeExam
	if err := r.DB.
		Preload("EntranceGlasses").
		Preload("EntranceContLens").
		Preload("UnaidedVADistance").
		Preload("UnaidedPHDistance").
		Preload("UnaidedVANear").
		Preload("AidedVADistance").
		Preload("AidedPHDistance").
		Preload("AidedVANear").
		Preload("Confrontation").
		Preload("Automated").
		Preload("Motility").
		Preload("Pupils").
		Preload("ColorVision").
		Preload("Bruckner").
		Preload("AmslerGrid").
		Preload("DistanceVonGraefePhoria").
		Preload("NearVonGraefePhoria").
		Preload("NearPointTesting").
		Preload("NearPointTesting.DistPhoria").
		Preload("NearPointTesting.NearPhoria").
		Preload("NearPointTesting.DistVergence").
		Preload("NearPointTesting.NearVergence").
		Preload("NearPointTesting.Accommodation").
		Preload("AutorefractorPreliminary").
		Preload("AutoKeratometerPreliminary").
		Preload("BloodPressure").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *PreliminaryEyeExamRepo) Create(eyeExamID int64) (*p.PreliminaryEyeExam, error) {
	v := p.PreliminaryEyeExam{EyeExamID: eyeExamID}
	if err := r.DB.Create(&v).Error; err != nil { return nil, err }
	return &v, nil
}

func (r *PreliminaryEyeExamRepo) Save(v *p.PreliminaryEyeExam) error { return r.DB.Save(v).Error }

// internal/repository/medical_repo/vision_exam_repo/chief_complaint/cc_hpi_eye.go
package chief_complaint

import (
	"errors"

	"gorm.io/gorm"
	cc "sighthub-backend/internal/models/medical/vision_exam/chief_complaint"
)

type CcHpiEyeRepo struct{ DB *gorm.DB }

func NewCcHpiEyeRepo(db *gorm.DB) *CcHpiEyeRepo {
	return &CcHpiEyeRepo{DB: db}
}

func (r *CcHpiEyeRepo) GetByEyeExamID(eyeExamID int64) (*cc.CcHpiEye, error) {
	var v cc.CcHpiEye
	if err := r.DB.
		Preload("ChiefComplaint").
		Preload("SecondaryComplaint").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

type CreateCcHpiEyeInput struct {
	EyeExamID                  int64
	ChiefComplaintHPIEyeID     *int64
	ChiefComplaintNote         *string
	SecondaryComplaintHPIEyeID *int64
}

func (r *CcHpiEyeRepo) Create(inp CreateCcHpiEyeInput) (*cc.CcHpiEye, error) {
	v := cc.CcHpiEye{
		EyeExamID:                  inp.EyeExamID,
		ChiefComplaintHPIEyeID:     inp.ChiefComplaintHPIEyeID,
		ChiefComplaintNote:         inp.ChiefComplaintNote,
		SecondaryComplaintHPIEyeID: inp.SecondaryComplaintHPIEyeID,
	}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *CcHpiEyeRepo) Save(v *cc.CcHpiEye) error {
	return r.DB.Save(v).Error
}

func (r *CcHpiEyeRepo) Delete(id int64) error {
	return r.DB.Delete(&cc.CcHpiEye{}, id).Error
}

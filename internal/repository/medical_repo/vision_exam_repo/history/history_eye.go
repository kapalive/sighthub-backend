// internal/repository/medical_repo/vision_exam_repo/history/history_eye.go
package history

import (
	"errors"

	"gorm.io/gorm"
	hmodel "sighthub-backend/internal/models/medical/vision_exam/history"
)

type HistoryEyeRepo struct{ DB *gorm.DB }

func NewHistoryEyeRepo(db *gorm.DB) *HistoryEyeRepo {
	return &HistoryEyeRepo{DB: db}
}

func (r *HistoryEyeRepo) GetByEyeExamID(eyeExamID int64) (*hmodel.HistoryEye, error) {
	var h hmodel.HistoryEye
	if err := r.DB.
		Preload("MedicalRecord").
		Preload("OcularHistory").
		Preload("ROSMedicalHistory").
		Preload("FamilyHistory").
		Preload("SocialHistory").
		Where("eye_exam_id = ?", eyeExamID).
		First(&h).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &h, nil
}

func (r *HistoryEyeRepo) GetByID(id int64) (*hmodel.HistoryEye, error) {
	var h hmodel.HistoryEye
	if err := r.DB.
		Preload("MedicalRecord").
		Preload("OcularHistory").
		Preload("ROSMedicalHistory").
		Preload("FamilyHistory").
		Preload("SocialHistory").
		First(&h, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &h, nil
}

type CreateHistoryEyeInput struct {
	MedicalRecordID     int64
	OcularHistoryID     int64
	ROSMedicalHistoryID int64
	FamilyHistoryID     int64
	SocialHistoryID     int64
	EyeExamID           int64
}

func (r *HistoryEyeRepo) Create(inp CreateHistoryEyeInput) (*hmodel.HistoryEye, error) {
	h := hmodel.HistoryEye{
		MedicalRecordID:     inp.MedicalRecordID,
		OcularHistoryID:     inp.OcularHistoryID,
		ROSMedicalHistoryID: inp.ROSMedicalHistoryID,
		FamilyHistoryID:     inp.FamilyHistoryID,
		SocialHistoryID:     inp.SocialHistoryID,
		EyeExamID:           inp.EyeExamID,
	}
	if err := r.DB.Create(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

type UpdateHistoryEyeInput struct {
	PrimaryCarePhysician       *string
	Other1PrimaryCarePhysician *string
	Other2PrimaryCarePhysician *string
	Medication                 *string
	Allergy                    *string
	NoMedications              *bool
	NoKnownAllergies           *bool
	LeadingWildcard            *bool
	SeeScannedDocumentsFolder  *bool
	HistoryNote                *string
}

func (r *HistoryEyeRepo) Update(id int64, inp UpdateHistoryEyeInput) error {
	updates := map[string]interface{}{}
	if inp.PrimaryCarePhysician != nil {
		updates["primary_care_physician"] = inp.PrimaryCarePhysician
	}
	if inp.Other1PrimaryCarePhysician != nil {
		updates["other_1_primary_care_physician"] = inp.Other1PrimaryCarePhysician
	}
	if inp.Other2PrimaryCarePhysician != nil {
		updates["other_2_primary_care_physician"] = inp.Other2PrimaryCarePhysician
	}
	if inp.Medication != nil {
		updates["medication"] = inp.Medication
	}
	if inp.Allergy != nil {
		updates["allergy"] = inp.Allergy
	}
	if inp.NoMedications != nil {
		updates["no_medications"] = *inp.NoMedications
	}
	if inp.NoKnownAllergies != nil {
		updates["no_known_allergies"] = *inp.NoKnownAllergies
	}
	if inp.LeadingWildcard != nil {
		updates["leading_wildcard"] = *inp.LeadingWildcard
	}
	if inp.SeeScannedDocumentsFolder != nil {
		updates["see_scanned_documents_folder"] = *inp.SeeScannedDocumentsFolder
	}
	if inp.HistoryNote != nil {
		updates["history_note"] = inp.HistoryNote
	}
	return r.DB.Model(&hmodel.HistoryEye{}).
		Where("id_history_eye = ?", id).Updates(updates).Error
}

package marketing_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/marketing"
)

type QuestionnaireReferralRepo struct{ DB *gorm.DB }

func NewQuestionnaireReferralRepo(db *gorm.DB) *QuestionnaireReferralRepo {
	return &QuestionnaireReferralRepo{DB: db}
}

func (r *QuestionnaireReferralRepo) GetByID(id int64) (*marketing.QuestionnaireReferral, error) {
	var qr marketing.QuestionnaireReferral
	if err := r.DB.Preload("VisitReason").Preload("ReferralSource").First(&qr, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &qr, nil
}

func (r *QuestionnaireReferralRepo) GetByPatientID(patientID int64) ([]marketing.QuestionnaireReferral, error) {
	var qrs []marketing.QuestionnaireReferral
	if err := r.DB.Preload("VisitReason").Preload("ReferralSource").
		Where("patient_id = ?", patientID).Find(&qrs).Error; err != nil {
		return nil, err
	}
	return qrs, nil
}

func (r *QuestionnaireReferralRepo) Create(qr *marketing.QuestionnaireReferral) error {
	return r.DB.Create(qr).Error
}

func (r *QuestionnaireReferralRepo) Save(qr *marketing.QuestionnaireReferral) error {
	return r.DB.Save(qr).Error
}

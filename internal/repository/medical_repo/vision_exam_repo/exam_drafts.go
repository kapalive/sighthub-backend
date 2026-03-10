// internal/repository/medical_repo/vision_exam_repo/exam_drafts.go
package vision_exam_repo

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vision_exam"
)

type ExamDraftsRepo struct{ DB *gorm.DB }

func NewExamDraftsRepo(db *gorm.DB) *ExamDraftsRepo {
	return &ExamDraftsRepo{DB: db}
}

func (r *ExamDraftsRepo) GetByPatientID(patientID int) ([]vision_exam.ExamDraft, error) {
	var list []vision_exam.ExamDraft
	if err := r.DB.Where("patient_id = ? AND completed = false", patientID).
		Order("updated_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ExamDraftsRepo) GetByID(id int) (*vision_exam.ExamDraft, error) {
	var d vision_exam.ExamDraft
	if err := r.DB.First(&d, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *ExamDraftsRepo) GetByExamID(examID string) (*vision_exam.ExamDraft, error) {
	var d vision_exam.ExamDraft
	if err := r.DB.Where("exam_id = ?", examID).First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *ExamDraftsRepo) Create(patientID int, examID string, data json.RawMessage) (*vision_exam.ExamDraft, error) {
	d := vision_exam.ExamDraft{
		PatientID: patientID,
		ExamID:    examID,
		Data:      data,
	}
	if err := r.DB.Create(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *ExamDraftsRepo) UpdateData(id int, data json.RawMessage) error {
	return r.DB.Model(&vision_exam.ExamDraft{}).
		Where("id_exam_draft = ?", id).
		Update("data", data).Error
}

func (r *ExamDraftsRepo) SetCompleted(id int) error {
	return r.DB.Model(&vision_exam.ExamDraft{}).
		Where("id_exam_draft = ?", id).
		Update("completed", true).Error
}

func (r *ExamDraftsRepo) Delete(id int) error {
	return r.DB.Delete(&vision_exam.ExamDraft{}, id).Error
}

// internal/repository/medical_repo/vision_exam_repo/exam_eye_note_doc.go
package vision_exam_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vision_exam"
)

type ExamEyeNoteDocRepo struct{ DB *gorm.DB }

func NewExamEyeNoteDocRepo(db *gorm.DB) *ExamEyeNoteDocRepo {
	return &ExamEyeNoteDocRepo{DB: db}
}

func (r *ExamEyeNoteDocRepo) GetByColID(colID int) (*vision_exam.ExamEyeNoteDoc, error) {
	var v vision_exam.ExamEyeNoteDoc
	if err := r.DB.Preload("Col").Where("exam_eye_note_col_id = ?", colID).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ExamEyeNoteDocRepo) Create(colID int, employeeID int64) (*vision_exam.ExamEyeNoteDoc, error) {
	v := vision_exam.ExamEyeNoteDoc{ExamEyeNoteColID: colID, EmployeeID: employeeID}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ExamEyeNoteDocRepo) Delete(id int64) error {
	return r.DB.Delete(&vision_exam.ExamEyeNoteDoc{}, id).Error
}

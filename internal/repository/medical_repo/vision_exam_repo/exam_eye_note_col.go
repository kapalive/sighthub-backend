// internal/repository/medical_repo/vision_exam_repo/exam_eye_note_col.go
package vision_exam_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/vision_exam"
)

type ExamEyeNoteColRepo struct{ DB *gorm.DB }

func NewExamEyeNoteColRepo(db *gorm.DB) *ExamEyeNoteColRepo {
	return &ExamEyeNoteColRepo{DB: db}
}

func (r *ExamEyeNoteColRepo) GetByTableID(tableID int) ([]vision_exam.ExamEyeNoteCol, error) {
	var list []vision_exam.ExamEyeNoteCol
	if err := r.DB.Preload("Table").Where("exam_eye_note_table_id = ?", tableID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ExamEyeNoteColRepo) GetByID(id int) (*vision_exam.ExamEyeNoteCol, error) {
	var v vision_exam.ExamEyeNoteCol
	if err := r.DB.Preload("Table").First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

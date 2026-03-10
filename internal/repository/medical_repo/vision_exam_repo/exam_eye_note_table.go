// internal/repository/medical_repo/vision_exam_repo/exam_eye_note_table.go
package vision_exam_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/vision_exam"
)

type ExamEyeNoteTableRepo struct{ DB *gorm.DB }

func NewExamEyeNoteTableRepo(db *gorm.DB) *ExamEyeNoteTableRepo {
	return &ExamEyeNoteTableRepo{DB: db}
}

func (r *ExamEyeNoteTableRepo) GetAll() ([]vision_exam.ExamEyeNoteTable, error) {
	var list []vision_exam.ExamEyeNoteTable
	if err := r.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ExamEyeNoteTableRepo) GetByID(id int) (*vision_exam.ExamEyeNoteTable, error) {
	var v vision_exam.ExamEyeNoteTable
	if err := r.DB.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

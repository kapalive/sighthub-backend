// internal/repository/medical_repo/vision_exam_repo/exam_eye_notes.go
package vision_exam_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/vision_exam"
)

type ExamEyeNotesRepo struct{ DB *gorm.DB }

func NewExamEyeNotesRepo(db *gorm.DB) *ExamEyeNotesRepo {
	return &ExamEyeNotesRepo{DB: db}
}

func (r *ExamEyeNotesRepo) GetByDocID(docID int64) ([]vision_exam.ExamEyeNotes, error) {
	var list []vision_exam.ExamEyeNotes
	if err := r.DB.Where("exam_eye_note_doc_id = ?", docID).Order("priority").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ExamEyeNotesRepo) Create(docID int64, note string, priority int) (*vision_exam.ExamEyeNotes, error) {
	v := vision_exam.ExamEyeNotes{ExamEyeNoteDocID: docID, Note: note, Priority: priority}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ExamEyeNotesRepo) Save(v *vision_exam.ExamEyeNotes) error {
	return r.DB.Save(v).Error
}

func (r *ExamEyeNotesRepo) Delete(id int64) error {
	return r.DB.Delete(&vision_exam.ExamEyeNotes{}, id).Error
}

func (r *ExamEyeNotesRepo) DeleteByDocID(docID int64) error {
	return r.DB.Where("exam_eye_note_doc_id = ?", docID).Delete(&vision_exam.ExamEyeNotes{}).Error
}

// internal/repository/medical_repo/vision_exam_repo/exam_result_eyes_files.go
package vision_exam_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/medical/vision_exam"
)

type ExamResultEyesFilesRepo struct{ DB *gorm.DB }

func NewExamResultEyesFilesRepo(db *gorm.DB) *ExamResultEyesFilesRepo {
	return &ExamResultEyesFilesRepo{DB: db}
}

func (r *ExamResultEyesFilesRepo) GetByID(id int64) (*vision_exam.ExamResultEyesFiles, error) {
	var v vision_exam.ExamResultEyesFiles
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ExamResultEyesFilesRepo) GetByPatientID(patientID int64) ([]vision_exam.ExamResultEyesFiles, error) {
	var items []vision_exam.ExamResultEyesFiles
	if err := r.DB.Where("patient_id = ?", patientID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ExamResultEyesFilesRepo) Create(v *vision_exam.ExamResultEyesFiles) error {
	return r.DB.Create(v).Error
}

func (r *ExamResultEyesFilesRepo) Save(v *vision_exam.ExamResultEyesFiles) error {
	return r.DB.Save(v).Error
}

func (r *ExamResultEyesFilesRepo) Delete(id int64) error {
	return r.DB.Delete(&vision_exam.ExamResultEyesFiles{}, id).Error
}

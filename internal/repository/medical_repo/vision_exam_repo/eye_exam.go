// internal/repository/medical_repo/vision_exam_repo/eye_exam.go
package vision_exam_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vision_exam"
)

type EyeExamRepo struct{ DB *gorm.DB }

func NewEyeExamRepo(db *gorm.DB) *EyeExamRepo {
	return &EyeExamRepo{DB: db}
}

func (r *EyeExamRepo) GetByPatientID(patientID int64) ([]vision_exam.EyeExam, error) {
	var list []vision_exam.EyeExam
	if err := r.DB.
		Preload("Employee").
		Preload("EyeExamType").
		Preload("Location").
		Where("patient_id = ?", patientID).
		Order("eye_exam_date DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *EyeExamRepo) GetByID(id int64) (*vision_exam.EyeExam, error) {
	var e vision_exam.EyeExam
	if err := r.DB.
		Preload("Employee").
		Preload("EyeExamType").
		Preload("Location").
		First(&e, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

type CreateEyeExamInput struct {
	EyeExamDate   time.Time
	EmployeeID    int64
	EyeExamTypeID int64
	LocationID    int
	PatientID     int64
}

func (r *EyeExamRepo) Create(inp CreateEyeExamInput) (*vision_exam.EyeExam, error) {
	e := vision_exam.EyeExam{
		EyeExamDate:   inp.EyeExamDate,
		EmployeeID:    inp.EmployeeID,
		EyeExamTypeID: inp.EyeExamTypeID,
		LocationID:    inp.LocationID,
		PatientID:     inp.PatientID,
	}
	if err := r.DB.Create(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EyeExamRepo) UpdateType(id, typeID int64) error {
	return r.DB.Model(&vision_exam.EyeExam{}).
		Where("id_eye_exam = ?", id).
		Update("eye_exam_type_id", typeID).Error
}

func (r *EyeExamRepo) SetPassed(id int64, passed bool) error {
	return r.DB.Model(&vision_exam.EyeExam{}).
		Where("id_eye_exam = ?", id).
		Update("passed", passed).Error
}

func (r *EyeExamRepo) Delete(id int64) error {
	return r.DB.Delete(&vision_exam.EyeExam{}, id).Error
}

// GetByLocationAndDate returns exams for a location on a given date.
func (r *EyeExamRepo) GetByLocationAndDate(locationID int, date time.Time) ([]vision_exam.EyeExam, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)
	var list []vision_exam.EyeExam
	if err := r.DB.
		Preload("Employee").
		Preload("EyeExamType").
		Where("location_id = ? AND eye_exam_date >= ? AND eye_exam_date < ?", locationID, start, end).
		Order("eye_exam_date").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

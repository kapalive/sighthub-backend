package vision_exam

import "time"

type ExamResultEyesFiles struct {
	IDExamResultEyesFiles int64     `gorm:"column:id_exam_result_eyes_files;primaryKey" json:"id_exam_result_eyes_files"`
	DateUpload            time.Time `gorm:"column:date_upload;not null;autoCreateTime" json:"date_upload"`
	PathToFile            string    `gorm:"column:path_to_file;size:255;not null" json:"path_to_file"`
	NameFile              string    `gorm:"column:name_file;size:255;not null" json:"name_file"`
	PatientID             int64     `gorm:"column:patient_id;not null" json:"patient_id"`
}

func (ExamResultEyesFiles) TableName() string { return "exam_result_eyes_files" }

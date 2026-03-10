package special

import "time"

type SpecialEyeFile struct {
	IDSpecialEyeFile  int64      `gorm:"column:id_special_eye_file;primaryKey;autoIncrement" json:"id_special_eye_file"`
	FilesUploadPath   *string    `gorm:"column:files_upload_path;size:255" json:"files_upload_path"`
	FileName          *string    `gorm:"column:file_name;size:255" json:"file_name"`
	DateRecord        *time.Time `gorm:"column:date_record;type:date" json:"date_record"`
	SpecialEyeExamID  int64      `gorm:"column:special_eye_exam_id;not null" json:"special_eye_exam_id"`
}

func (SpecialEyeFile) TableName() string { return "special_eye_file" }

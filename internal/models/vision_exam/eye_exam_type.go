// internal/models/vision_exam/eye_exam_type.go
package vision_exam

// EyeExamType ↔ table: eye_exam_type
type EyeExamType struct {
	IDEyeExamType int64  `gorm:"column:id_eye_exam_type;primaryKey;autoIncrement" json:"id_eye_exam_type"`
	ExamTypeName  string `gorm:"column:exam_type_name;type:text;not null"         json:"exam_type_name"`
}

func (EyeExamType) TableName() string { return "eye_exam_type" }

func (e *EyeExamType) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_eye_exam_type": e.IDEyeExamType,
		"exam_type_name":   e.ExamTypeName,
	}
}

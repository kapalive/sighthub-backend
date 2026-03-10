package special

type SpecialEyeExam struct {
	IDSpecialEyeExam int64   `gorm:"column:id_special_eye_exam;primaryKey;autoIncrement" json:"id_special_eye_exam"`
	SpecialTesting   *string `gorm:"column:special_testing;type:text" json:"special_testing"`
	EyeExamID        int64   `gorm:"column:eye_exam_id;not null" json:"eye_exam_id"`

	Files []SpecialEyeFile `gorm:"foreignKey:SpecialEyeExamID" json:"files"`
}

func (SpecialEyeExam) TableName() string { return "special_eye_exam" }

package vision_exam

type ExamEyeNoteTable struct {
	IDExamEyeNoteTable int    `gorm:"column:id_exam_eye_note_table;primaryKey;autoIncrement" json:"id_exam_eye_note_table"`
	Name               string `gorm:"column:table_name;type:varchar(100);not null"           json:"table_name"`
}

func (ExamEyeNoteTable) TableName() string { return "exam_eye_note_table" }

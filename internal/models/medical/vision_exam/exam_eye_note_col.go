package vision_exam

type ExamEyeNoteCol struct {
	IDExamEyeNoteCol   int    `gorm:"column:id_exam_eye_note_col;primaryKey;autoIncrement" json:"id_exam_eye_note_col"`
	ColumnName         string `gorm:"column:column_name;type:varchar(100);not null"        json:"column_name"`
	ExamEyeNoteTableID int    `gorm:"column:exam_eye_note_table_id;not null"               json:"exam_eye_note_table_id"`
}

func (ExamEyeNoteCol) TableName() string { return "exam_eye_note_col" }

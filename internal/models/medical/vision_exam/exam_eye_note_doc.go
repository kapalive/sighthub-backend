package vision_exam

type ExamEyeNoteDoc struct {
	IDExamEyeNoteDoc int64 `gorm:"column:id_exam_eye_note_doc;primaryKey;autoIncrement" json:"id_exam_eye_note_doc"`
	ExamEyeNoteColID int   `gorm:"column:exam_eye_note_col_id;not null"                json:"exam_eye_note_col_id"`
	EmployeeID       int64 `gorm:"column:employee_id;not null"                         json:"employee_id"`
}

func (ExamEyeNoteDoc) TableName() string { return "exam_eye_note_doc" }

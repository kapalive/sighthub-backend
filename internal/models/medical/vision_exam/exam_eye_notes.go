package vision_exam

type ExamEyeNotes struct {
	IDExamEyeNotes   int64  `gorm:"column:id_exam_eye_notes;primaryKey;autoIncrement" json:"id_exam_eye_notes"`
	ExamEyeNoteDocID int64  `gorm:"column:exam_eye_note_doc_id;not null"              json:"exam_eye_note_doc_id"`
	Note             string `gorm:"column:note;type:text;not null"                    json:"note"`
	Priority         int    `gorm:"column:priority;default:0"                         json:"priority"`
}

func (ExamEyeNotes) TableName() string { return "exam_eye_notes" }

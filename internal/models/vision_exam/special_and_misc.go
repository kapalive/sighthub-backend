// internal/models/vision_exam/special_and_misc.go
package vision_exam

import (
	"time"
	"sighthub-backend/internal/models/invoices"
)

// SpecialEyeExam ↔ table: special_eye_exam
type SpecialEyeExam struct {
	IDSpecialEyeExam int64   `gorm:"column:id_special_eye_exam;primaryKey;autoIncrement" json:"id_special_eye_exam"`
	SpecialTesting   *string `gorm:"column:special_testing;type:text"                    json:"special_testing,omitempty"`
	EyeExamID        int64   `gorm:"column:eye_exam_id;not null"                         json:"eye_exam_id"`

	Files []SpecialEyeFile `gorm:"foreignKey:SpecialEyeExamID;references:IDSpecialEyeExam" json:"-"`
}
func (SpecialEyeExam) TableName() string { return "special_eye_exam" }
func (s *SpecialEyeExam) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_special_eye_exam": s.IDSpecialEyeExam, "special_testing": s.SpecialTesting, "eye_exam_id": s.EyeExamID,
	}
}

// SpecialEyeFile ↔ table: special_eye_file
type SpecialEyeFile struct {
	IDSpecialEyeFile  int64      `gorm:"column:id_special_eye_file;primaryKey;autoIncrement" json:"id_special_eye_file"`
	FilesUploadPath   *string    `gorm:"column:files_upload_path;type:varchar(255)"          json:"files_upload_path,omitempty"`
	FileName          *string    `gorm:"column:file_name;type:varchar(255)"                  json:"file_name,omitempty"`
	DateRecord        *time.Time `gorm:"column:date_record;type:date"                        json:"date_record,omitempty"`
	SpecialEyeExamID  int64      `gorm:"column:special_eye_exam_id;not null"                 json:"special_eye_exam_id"`
}
func (SpecialEyeFile) TableName() string { return "special_eye_file" }
func (s *SpecialEyeFile) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_special_eye_file": s.IDSpecialEyeFile, "files_upload_path": s.FilesUploadPath,
		"file_name": s.FileName, "special_eye_exam_id": s.SpecialEyeExamID,
	}
	if s.DateRecord != nil { m["date_record"] = s.DateRecord.Format("2006-01-02") } else { m["date_record"] = nil }
	return m
}

// SuperEyeExam ↔ table: super_eye_exam (links exam to invoice)
type SuperEyeExam struct {
	IDSuperEyeExam int64  `gorm:"column:id_super_eye_exam;primaryKey;autoIncrement"   json:"id_super_eye_exam"`
	InvoiceID      *int64 `gorm:"column:invoice_id"                                   json:"invoice_id,omitempty"`
	EyeExamID      int64  `gorm:"column:eye_exam_id;not null;uniqueIndex"             json:"eye_exam_id"`

	Invoice *invoices.Invoice `gorm:"foreignKey:InvoiceID;references:IDInvoice" json:"-"`
}
func (SuperEyeExam) TableName() string { return "super_eye_exam" }
func (s *SuperEyeExam) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_super_eye_exam": s.IDSuperEyeExam, "invoice_id": s.InvoiceID, "eye_exam_id": s.EyeExamID,
	}
}

// ExamEyeNoteTable ↔ table: exam_eye_note_table
type ExamEyeNoteTable struct {
	IDExamEyeNoteTable int    `gorm:"column:id_exam_eye_note_table;primaryKey;autoIncrement" json:"id_exam_eye_note_table"`
	TableName_         string `gorm:"column:table_name;type:varchar(100);not null"           json:"table_name"`
}
func (ExamEyeNoteTable) TableName() string { return "exam_eye_note_table" }
func (e *ExamEyeNoteTable) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_exam_eye_note_table": e.IDExamEyeNoteTable, "table_name": e.TableName_}
}

// ExamEyeNoteCol ↔ table: exam_eye_note_col
type ExamEyeNoteCol struct {
	IDExamEyeNoteCol    int    `gorm:"column:id_exam_eye_note_col;primaryKey;autoIncrement" json:"id_exam_eye_note_col"`
	ColumnName          string `gorm:"column:column_name;type:varchar(100);not null"        json:"column_name"`
	ExamEyeNoteTableID  int    `gorm:"column:exam_eye_note_table_id;not null"               json:"exam_eye_note_table_id"`

	Table *ExamEyeNoteTable `gorm:"foreignKey:ExamEyeNoteTableID;references:IDExamEyeNoteTable" json:"-"`
}
func (ExamEyeNoteCol) TableName() string { return "exam_eye_note_col" }
func (e *ExamEyeNoteCol) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_exam_eye_note_col": e.IDExamEyeNoteCol, "column_name": e.ColumnName, "exam_eye_note_table_id": e.ExamEyeNoteTableID}
}

// ExamEyeNoteDoc ↔ table: exam_eye_note_doc
type ExamEyeNoteDoc struct {
	IDExamEyeNoteDoc   int64 `gorm:"column:id_exam_eye_note_doc;primaryKey;autoIncrement" json:"id_exam_eye_note_doc"`
	ExamEyeNoteColID   int   `gorm:"column:exam_eye_note_col_id;not null"                 json:"exam_eye_note_col_id"`
	EmployeeID         int64 `gorm:"column:employee_id;not null"                          json:"employee_id"`

	Col *ExamEyeNoteCol `gorm:"foreignKey:ExamEyeNoteColID;references:IDExamEyeNoteCol" json:"-"`
}
func (ExamEyeNoteDoc) TableName() string { return "exam_eye_note_doc" }
func (e *ExamEyeNoteDoc) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_exam_eye_note_doc": e.IDExamEyeNoteDoc, "exam_eye_note_col_id": e.ExamEyeNoteColID, "employee_id": e.EmployeeID}
}

// ExamEyeNotes ↔ table: exam_eye_notes
type ExamEyeNotes struct {
	IDExamEyeNotes    int64  `gorm:"column:id_exam_eye_notes;primaryKey;autoIncrement" json:"id_exam_eye_notes"`
	ExamEyeNoteDocID  int64  `gorm:"column:exam_eye_note_doc_id;not null"              json:"exam_eye_note_doc_id"`
	Note              string `gorm:"column:note;type:text;not null"                    json:"note"`
	Priority          int    `gorm:"column:priority;not null;default:0"                json:"priority"`
}
func (ExamEyeNotes) TableName() string { return "exam_eye_notes" }
func (e *ExamEyeNotes) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_exam_eye_notes": e.IDExamEyeNotes, "exam_eye_note_doc_id": e.ExamEyeNoteDocID, "note": e.Note, "priority": e.Priority}
}

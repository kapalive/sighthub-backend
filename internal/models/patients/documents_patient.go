package patients

import "time"

type DocumentsPatient struct {
	IDDocumentsPatient int64      `gorm:"column:id_documents_patient;primaryKey;autoIncrement" json:"id_documents_patient"`
	PatientID          int64      `gorm:"column:patient_id;not null"                           json:"patient_id"`
	FileName           string     `gorm:"column:file_name;not null"                            json:"file_name"`
	FilePath           string     `gorm:"column:file_path;not null"                            json:"file_path"`
	DocumentType       *string    `gorm:"column:document_type"                                 json:"document_type,omitempty"`
	Description        *string    `gorm:"column:description;type:text"                         json:"description,omitempty"`
	CreatedBy          *int64     `gorm:"column:created_by"                                    json:"created_by,omitempty"`
	CreatedTime        *time.Time `gorm:"column:created_time;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_time,omitempty"`
	IsHidden           bool       `gorm:"column:is_hidden;default:false"                       json:"is_hidden"`
}

func (DocumentsPatient) TableName() string { return "documents_patient" }

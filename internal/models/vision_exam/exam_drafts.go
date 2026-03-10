// internal/models/vision_exam/exam_drafts.go
package vision_exam

import (
	"encoding/json"
	"time"
)

// ExamDraft ↔ table: exam_drafts
type ExamDraft struct {
	IDExamDraft int       `gorm:"column:id_exam_draft;primaryKey;autoIncrement" json:"id_exam_draft"`
	PatientID   int       `gorm:"column:patient_id;not null"                    json:"patient_id"`
	StartedAt   time.Time `gorm:"column:started_at;type:timestamptz;not null;autoCreateTime" json:"started_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime" json:"updated_at"`
	Completed   bool      `gorm:"column:completed;not null;default:false"        json:"completed"`
	Data        json.RawMessage `gorm:"column:data;type:jsonb;not null;default:'{}'" json:"data"`
	ExamID      string    `gorm:"column:exam_id;not null"                        json:"exam_id"`
}

func (ExamDraft) TableName() string { return "exam_drafts" }

func (e *ExamDraft) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_exam_draft": e.IDExamDraft,
		"patient_id":    e.PatientID,
		"started_at":    e.StartedAt.Format(time.RFC3339),
		"updated_at":    e.UpdatedAt.Format(time.RFC3339),
		"completed":     e.Completed,
		"data":          e.Data,
		"exam_id":       e.ExamID,
	}
}

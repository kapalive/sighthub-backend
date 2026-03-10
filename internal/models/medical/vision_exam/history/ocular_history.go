// internal/models/medical/vision_exam/history/ocular_history.go
package history

// OcularHistory ↔ table: ocular_history
type OcularHistory struct {
	IDOcularHistory    int64   `gorm:"column:id_ocular_history;primaryKey;autoIncrement"  json:"id_ocular_history"`
	PrevEyeHistoryNone *bool   `gorm:"column:prev_eye_history_none"                       json:"prev_eye_history_none,omitempty"`
	PrevEyeHistory     *string `gorm:"column:prev_eye_history;type:text"                  json:"prev_eye_history,omitempty"`
	LastExamNone       *bool   `gorm:"column:last_exam_none"                              json:"last_exam_none,omitempty"`
	LastExam           *string `gorm:"column:last_exam;type:text"                         json:"last_exam,omitempty"`
	ClCurrentWearNone  *bool   `gorm:"column:cl_current_wear_none"                        json:"cl_current_wear_none,omitempty"`
	ClCurrentWearScl   *bool   `gorm:"column:cl_current_wear_scl"                         json:"cl_current_wear_scl,omitempty"`
	ClCurrentWearRgp   *bool   `gorm:"column:cl_current_wear_rgp"                         json:"cl_current_wear_rgp,omitempty"`
	ClCurrentWearOther *string `gorm:"column:cl_current_wear_other;type:text"             json:"cl_current_wear_other,omitempty"`
	ModalityDaily      *bool   `gorm:"column:modality_daily"                              json:"modality_daily,omitempty"`
	ModalityBiweekly   *bool   `gorm:"column:modality_biweekly"                           json:"modality_biweekly,omitempty"`
	ModalityMonthly    *bool   `gorm:"column:modality_monthly"                            json:"modality_monthly,omitempty"`
	ModalityAnnually   *bool   `gorm:"column:modality_annually"                           json:"modality_annually,omitempty"`
	ModalityOther      *string `gorm:"column:modality_other;type:text"                    json:"modality_other,omitempty"`
	ModalitySolutions  *string `gorm:"column:modality_solutions;type:text"                json:"modality_solutions,omitempty"`
}

func (OcularHistory) TableName() string { return "ocular_history" }

func (o *OcularHistory) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_ocular_history":     o.IDOcularHistory,
		"prev_eye_history_none": o.PrevEyeHistoryNone,
		"prev_eye_history":      o.PrevEyeHistory,
		"last_exam_none":        o.LastExamNone,
		"last_exam":             o.LastExam,
		"cl_current_wear_none":  o.ClCurrentWearNone,
		"cl_current_wear_scl":   o.ClCurrentWearScl,
		"cl_current_wear_rgp":   o.ClCurrentWearRgp,
		"cl_current_wear_other": o.ClCurrentWearOther,
		"modality_daily":        o.ModalityDaily,
		"modality_biweekly":     o.ModalityBiweekly,
		"modality_monthly":      o.ModalityMonthly,
		"modality_annually":     o.ModalityAnnually,
		"modality_other":        o.ModalityOther,
		"modality_solutions":    o.ModalitySolutions,
	}
}

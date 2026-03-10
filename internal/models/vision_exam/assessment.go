// internal/models/vision_exam/assessment.go
package vision_exam

import "sighthub-backend/internal/models/general"

// AssessmentEye ↔ table: assessment_eye
type AssessmentEye struct {
	IDAssessmentEye int64   `gorm:"column:id_assessment_eye;primaryKey;autoIncrement" json:"id_assessment_eye"`
	EyeExamID       int64   `gorm:"column:eye_exam_id;not null"                       json:"eye_exam_id"`
	Plan            *string `gorm:"column:plan;type:text"                             json:"plan,omitempty"`
	Impression      *string `gorm:"column:impression;type:varchar(255)"               json:"impression,omitempty"`

	Diagnoses []AssessmentDiagnosisEye `gorm:"foreignKey:AssessmentEyeID;references:IDAssessmentEye" json:"-"`
	PQRSItems []AssessmentPQRS         `gorm:"foreignKey:AssessmentEyeID;references:IDAssessmentEye" json:"-"`
}
func (AssessmentEye) TableName() string { return "assessment_eye" }
func (a *AssessmentEye) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_assessment_eye": a.IDAssessmentEye, "eye_exam_id": a.EyeExamID,
		"plan": a.Plan, "impression": a.Impression,
	}
	return m
}

// AssessmentDiagnosisEye ↔ table: assessment_diagnosis_eye
type AssessmentDiagnosisEye struct {
	IDAssessmentDiagnosis int64   `gorm:"column:id_assessment_diagnosis;primaryKey;autoIncrement" json:"id_assessment_diagnosis"`
	AssessmentEyeID       int64   `gorm:"column:assessment_eye_id;not null"                       json:"assessment_eye_id"`
	Code                  string  `gorm:"column:code;type:varchar(10);not null"                   json:"code"`
	LevelID               int64   `gorm:"column:level_id;not null"                                json:"level_id"`
	Type                  string  `gorm:"column:type;type:varchar(50);not null"                   json:"type"`
	Title                 *string `gorm:"column:title;type:varchar(255)"                          json:"title,omitempty"`
}
func (AssessmentDiagnosisEye) TableName() string { return "assessment_diagnosis_eye" }
func (a *AssessmentDiagnosisEye) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_assessment_diagnosis": a.IDAssessmentDiagnosis,
		"assessment_eye_id": a.AssessmentEyeID, "code": a.Code,
		"level_id": a.LevelID, "type": a.Type, "title": a.Title,
	}
}

// AssessmentPQRS ↔ table: assessment_pqrs (link assessment ↔ pqrs)
type AssessmentPQRS struct {
	IDAssessmentPQRS int64 `gorm:"column:id_assessment_pqrs;primaryKey;autoIncrement" json:"id_assessment_pqrs"`
	AssessmentEyeID  int64 `gorm:"column:assessment_eye_id;not null"                  json:"assessment_eye_id"`
	PQRSid           int64 `gorm:"column:pqrs_id;not null"                            json:"pqrs_id"`

	PQRS *general.PQRS `gorm:"foreignKey:PQRSid;references:IDPQRS" json:"-"`
}
func (AssessmentPQRS) TableName() string { return "assessment_pqrs" }
func (a *AssessmentPQRS) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_assessment_pqrs": a.IDAssessmentPQRS,
		"assessment_eye_id": a.AssessmentEyeID, "pqrs_id": a.PQRSid,
	}
	if a.PQRS != nil { m["pqrs"] = a.PQRS.ToMap() }
	return m
}

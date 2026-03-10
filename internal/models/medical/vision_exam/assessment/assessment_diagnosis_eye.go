package assessment

type AssessmentDiagnosis struct {
	IDAssessmentDiagnosis int64   `gorm:"column:id_assessment_diagnosis;primaryKey;autoIncrement" json:"id_assessment_diagnosis"`
	AssessmentEyeID       int64   `gorm:"column:assessment_eye_id;not null" json:"assessment_eye_id"`
	Code                  *string `gorm:"column:code;size:10" json:"code"`
	LevelID               *int64  `gorm:"column:level_id" json:"level_id"`
	Type                  *string `gorm:"column:type;size:50" json:"type"`
	Title                 *string `gorm:"column:title;size:255" json:"title"`
}

func (AssessmentDiagnosis) TableName() string { return "assessment_diagnosis_eye" }

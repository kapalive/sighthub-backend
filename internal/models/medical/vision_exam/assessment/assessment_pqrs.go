package assessment

type AssessmentPQRS struct {
	IDAssessmentPQRS int64 `gorm:"column:id_assessment_pqrs;primaryKey;autoIncrement" json:"id_assessment_pqrs"`
	AssessmentEyeID  int64 `gorm:"column:assessment_eye_id;not null" json:"assessment_eye_id"`
	PqrsID           int64 `gorm:"column:pqrs_id;not null" json:"pqrs_id"`
}

func (AssessmentPQRS) TableName() string { return "assessment_pqrs" }

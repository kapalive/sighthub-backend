package assessment

type AssessmentEye struct {
	IDAssessmentEye int64  `gorm:"column:id_assessment_eye;primaryKey;autoIncrement" json:"id_assessment_eye"`
	EyeExamID       int64  `gorm:"column:eye_exam_id;not null" json:"eye_exam_id"`
	Plan            *string `gorm:"column:plan;type:text" json:"plan"`
	Impression      *string `gorm:"column:impression;size:255" json:"impression"`

	Diagnoses []AssessmentDiagnosis `gorm:"foreignKey:AssessmentEyeID" json:"diagnoses"`
	PQRSItems []AssessmentPQRS      `gorm:"foreignKey:AssessmentEyeID" json:"pqrs"`
}

func (AssessmentEye) TableName() string { return "assessment_eye" }

package slrp

type AssessmentSLRPEyeExam struct {
	IDAssessmentSLRPEyeExam int64   `gorm:"column:id_assessment_slrp_eye_exam;primaryKey;autoIncrement" json:"id_assessment_slrp_eye_exam"`
	ToleratedSessionWellToday *bool `gorm:"column:tolerated_session_well_today;default:false" json:"tolerated_session_well_today"`
	Comments                  *string `gorm:"column:comments;type:text" json:"comments"`
}

func (AssessmentSLRPEyeExam) TableName() string { return "assessment_slrp_eye_exam" }

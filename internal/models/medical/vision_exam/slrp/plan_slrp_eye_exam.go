package slrp

type PlanSLRPEyeExam struct {
	IDPlanSLRPEyeExam        int64   `gorm:"column:id_plan_slrp_eye_exam;primaryKey;autoIncrement" json:"id_plan_slrp_eye_exam"`
	ContinueProgramScheduled *bool   `gorm:"column:continue_program_scheduled;default:false" json:"continue_program_scheduled"`
	ModifyProgram            *string `gorm:"column:modify_program;type:text" json:"modify_program"`
}

func (PlanSLRPEyeExam) TableName() string { return "plan_slrp_eye_exam" }

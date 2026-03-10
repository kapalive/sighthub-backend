package slrp

import "time"

type SLRPEyeExam struct {
	IDSLRPEyeExam              int64      `gorm:"column:id_slrp_eye_exam;primaryKey;autoIncrement" json:"id_slrp_eye_exam"`
	EyeExamID                  int64      `gorm:"column:eye_exam_id;not null" json:"eye_exam_id"`
	SubjectiveSLRPEyeExamID    *int64     `gorm:"column:subjective_slrp_eye_exam_id" json:"subjective_slrp_eye_exam_id"`
	ObjectiveSLRPEyeExamID     *int64     `gorm:"column:objective_slrp_eye_exam_id" json:"objective_slrp_eye_exam_id"`
	AssessmentSLRPEyeExamID    *int64     `gorm:"column:assessment_slrp_eye_exam_id" json:"assessment_slrp_eye_exam_id"`
	PlanSLRPEyeExamID          *int64     `gorm:"column:plan_slrp_eye_exam_id" json:"plan_slrp_eye_exam_id"`
	StartDate                  *time.Time `gorm:"column:start_date;type:date" json:"start_date"`
	StartTime                  *string    `gorm:"column:start_time;type:time" json:"start_time"`
	EndDate                    *time.Time `gorm:"column:end_date;type:date" json:"end_date"`
	EndTime                    *string    `gorm:"column:end_time;type:time" json:"end_time"`

	Subjective  *SubjectiveSLRPEyeExam  `gorm:"foreignKey:IDSubjectiveSLRPEyeExam;references:SubjectiveSLRPEyeExamID" json:"subjective"`
	Objective   *ObjectiveSLRPEyeExam   `gorm:"foreignKey:IDObjectiveSLRPEyeExam;references:ObjectiveSLRPEyeExamID" json:"objective"`
	Assessment  *AssessmentSLRPEyeExam  `gorm:"foreignKey:IDAssessmentSLRPEyeExam;references:AssessmentSLRPEyeExamID" json:"assessment"`
	Plan        *PlanSLRPEyeExam        `gorm:"foreignKey:IDPlanSLRPEyeExam;references:PlanSLRPEyeExamID" json:"plan"`
}

func (SLRPEyeExam) TableName() string { return "slrp_eye_exam" }

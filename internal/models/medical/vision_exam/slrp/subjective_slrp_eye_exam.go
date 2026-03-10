package slrp

type SubjectiveSLRPEyeExam struct {
	IDSubjectiveSLRPEyeExam int64   `gorm:"column:id_subjective_slrp_eye_exam;primaryKey;autoIncrement" json:"id_subjective_slrp_eye_exam"`
	AccompaniedBy           *string `gorm:"column:accompanied_by" json:"accompanied_by"`
	SupervisingTechnician   *string `gorm:"column:supervising_technician" json:"supervising_technician"`
	Changes                 *string `gorm:"column:changes" json:"changes"`
}

func (SubjectiveSLRPEyeExam) TableName() string { return "subjective_slrp_eye_exam" }

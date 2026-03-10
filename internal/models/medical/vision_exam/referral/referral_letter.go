package referral

type ReferralLetter struct {
	IDReferralLetter   int64   `gorm:"column:id_referral_letter;primaryKey" json:"id_referral_letter"`
	TitleLetter        *string `gorm:"column:title_letter" json:"title_letter"`
	IntroLetter        *string `gorm:"column:intro_letter;type:text" json:"intro_letter"`
	TestsLetter        *string `gorm:"column:tests_letter;type:text" json:"tests_letter"`
	IssueLetter        *string `gorm:"column:issue_letter;type:text" json:"issue_letter"`
	EyeExamID          int64   `gorm:"column:eye_exam_id;not null" json:"eye_exam_id"`
	ToReferralDoctorID *int64  `gorm:"column:to_referral_doctor_id" json:"to_referral_doctor_id"`
	CcReferralDoctorID *int64  `gorm:"column:cc_referral_doctor_id" json:"cc_referral_doctor_id"`

	ToDoctor *ReferralDoctor `gorm:"foreignKey:ToReferralDoctorID" json:"to_doctor"`
	CcDoctor *ReferralDoctor `gorm:"foreignKey:CcReferralDoctorID" json:"cc_doctor"`
}

func (ReferralLetter) TableName() string { return "referral_letter" }

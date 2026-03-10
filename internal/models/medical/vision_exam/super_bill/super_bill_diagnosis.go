package super_bill

type SuperBillDiagnosis struct {
	IDSuperBillDiagnosis  int64 `gorm:"column:id_super_bill_diagnosis;primaryKey;autoIncrement" json:"id_super_bill_diagnosis"`
	SuperEyeExamID        int64 `gorm:"column:super_eye_exam_id;not null" json:"super_eye_exam_id"`
	ProfessionalServiceID int64 `gorm:"column:professional_service_id;not null" json:"professional_service_id"`
}

func (SuperBillDiagnosis) TableName() string { return "super_bill_diagnosis" }

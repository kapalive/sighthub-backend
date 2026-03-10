package super_bill

type SuperEyeExam struct {
	IDSuperEyeExam int64  `gorm:"column:id_super_eye_exam;primaryKey;autoIncrement" json:"id_super_eye_exam"`
	InvoiceID      *int64 `gorm:"column:invoice_id" json:"invoice_id"`
	EyeExamID      int64  `gorm:"column:eye_exam_id;not null;uniqueIndex" json:"eye_exam_id"`

	Diagnoses      []SuperBillDiagnosis `gorm:"foreignKey:SuperEyeExamID" json:"diagnoses"`
	DiseaseBills   []DiseaseSuperBill   `gorm:"foreignKey:SuperBillDiagnosisID" json:"disease_bills"`
}

func (SuperEyeExam) TableName() string { return "super_eye_exam" }

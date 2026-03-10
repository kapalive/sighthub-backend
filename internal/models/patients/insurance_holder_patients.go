package patients

type InsuranceHolderPatients struct {
	IDInsuranceHolderPatients int    `gorm:"column:id_insurance_holder_patients;primaryKey;autoIncrement" json:"id_insurance_holder_patients"`
	InsurancePolicyID         int64  `gorm:"column:insurance_policy_id;not null;uniqueIndex:uix_policy_patient" json:"insurance_policy_id"`
	PatientID                 int64  `gorm:"column:patient_id;not null;uniqueIndex:uix_policy_patient"          json:"patient_id"`
	HolderType                string `gorm:"column:holder_type;size:50;not null"                                json:"holder_type"`
	Position                  *string `gorm:"column:position;size:50"                                           json:"position,omitempty"`
	MemberNumber               *string `gorm:"column:member_number;size:50"                                     json:"member_number,omitempty"`
}

func (InsuranceHolderPatients) TableName() string { return "insurance_holder_patients" }

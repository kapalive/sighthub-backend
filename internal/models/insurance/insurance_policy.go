package insurance

import "time"

type InsurancePolicy struct {
	IDInsurancePolicy       int64   `gorm:"column:id_insurance_policy;primaryKey"      json:"id_insurance_policy"`
	GroupNumber             *string `gorm:"column:group_number;type:varchar(50)"        json:"group_number,omitempty"`
	CoverageDetails         *string `gorm:"column:coverage_details;type:text"           json:"coverage_details,omitempty"`
	Specify                 *string `gorm:"column:specify;type:varchar(255)"            json:"specify,omitempty"`
	Active                  bool    `gorm:"column:active;not null;default:true"         json:"active"`
	InsurancePolicyFilePath *string `gorm:"column:insurance_policy_file_path;type:varchar(255)" json:"insurance_policy_file_path,omitempty"`

	InsuranceCompanyID      int  `gorm:"column:insurance_company_id;not null"         json:"insurance_company_id"`
	InsuranceCoverageTypeID *int `gorm:"column:insurance_coverage_type_id"            json:"insurance_coverage_type_id,omitempty"`

	FrontPhoto *string    `gorm:"column:front_photo;type:varchar(255)"                 json:"front_photo,omitempty"`
	BackPhoto  *string    `gorm:"column:back_photo;type:varchar(255)"                  json:"back_photo,omitempty"`
	DateUpload *time.Time `gorm:"column:date_upload;type:timestamptz"                  json:"-"`

	// Relations
	InsuranceCompany      *InsuranceCompany      `gorm:"foreignKey:InsuranceCompanyID;references:IDInsuranceCompany"                json:"-"`
	InsuranceCoverageType *InsuranceCoverageType `gorm:"foreignKey:InsuranceCoverageTypeID;references:IDInsuranceCoverageType"      json:"-"`
}

func (InsurancePolicy) TableName() string { return "insurance_policy" }

func (p *InsurancePolicy) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_insurance_policy":        p.IDInsurancePolicy,
		"group_number":               p.GroupNumber,
		"coverage_details":           p.CoverageDetails,
		"specify":                    p.Specify,
		"active":                     p.Active,
		"insurance_policy_file_path": p.InsurancePolicyFilePath,
		"insurance_company_id":       p.InsuranceCompanyID,
		"insurance_coverage_type_id": p.InsuranceCoverageTypeID,
		"front_photo":                p.FrontPhoto,
		"back_photo":                 p.BackPhoto,
	}

	if p.DateUpload != nil && !p.DateUpload.IsZero() {
		m["date_upload"] = p.DateUpload.Format(time.RFC3339)
	} else {
		m["date_upload"] = nil
	}

	return m
}

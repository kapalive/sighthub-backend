// internal/models/insurance/claim_template.go
package insurance

import "time"

// ClaimTemplate ⇄ claim_template
type ClaimTemplate struct {
	IDClaimTemplate          int        `gorm:"column:id_claim_template;primaryKey;autoIncrement" json:"id_claim_template"`
	Name                     string     `gorm:"column:name;type:varchar(255);not null"             json:"name"`
	Description              *string    `gorm:"column:description;type:text"                       json:"description,omitempty"`
	LocationID               *int       `gorm:"column:location_id"                                 json:"location_id,omitempty"`
	DoctorID                 *int       `gorm:"column:doctor_id"                                   json:"doctor_id,omitempty"`
	InsuranceCompanyID       *int       `gorm:"column:insurance_company_id"                        json:"insurance_company_id,omitempty"`
	InsuranceType            *string    `gorm:"column:insurance_type;type:varchar(20)"             json:"insurance_type,omitempty"`
	InsuranceName            *string    `gorm:"column:insurance_name;type:varchar(255)"            json:"insurance_name,omitempty"`
	InsuranceAddress         *string    `gorm:"column:insurance_address;type:varchar(255)"         json:"insurance_address,omitempty"`
	InsuranceAddress2        *string    `gorm:"column:insurance_address2;type:varchar(255)"        json:"insurance_address2,omitempty"`
	InsuranceCityStateZip    *string    `gorm:"column:insurance_city_state_zip;type:varchar(255)"  json:"insurance_city_state_zip,omitempty"`
	ConditionEmployment      *string    `gorm:"column:condition_employment;type:varchar(10)"       json:"condition_employment,omitempty"`
	ConditionAutoAccident    *string    `gorm:"column:condition_auto_accident;type:varchar(10)"    json:"condition_auto_accident,omitempty"`
	AccidentState            *string    `gorm:"column:accident_state;type:varchar(10)"             json:"accident_state,omitempty"`
	ConditionOtherAccident   *string    `gorm:"column:condition_other_accident;type:varchar(10)"   json:"condition_other_accident,omitempty"`
	PatientSignature         *string    `gorm:"column:patient_signature;type:varchar(255)"         json:"patient_signature,omitempty"`
	InsuredSignature         *string    `gorm:"column:insured_signature;type:varchar(255)"         json:"insured_signature,omitempty"`
	Service1RenderingNpi     *string    `gorm:"column:service1_rendering_npi;type:varchar(50)"     json:"service1_rendering_npi,omitempty"`
	Service2RenderingNpi     *string    `gorm:"column:service2_rendering_npi;type:varchar(50)"     json:"service2_rendering_npi,omitempty"`
	Service3RenderingNpi     *string    `gorm:"column:service3_rendering_npi;type:varchar(50)"     json:"service3_rendering_npi,omitempty"`
	Service4RenderingNpi     *string    `gorm:"column:service4_rendering_npi;type:varchar(50)"     json:"service4_rendering_npi,omitempty"`
	Service5RenderingNpi     *string    `gorm:"column:service5_rendering_npi;type:varchar(50)"     json:"service5_rendering_npi,omitempty"`
	Service6RenderingNpi     *string    `gorm:"column:service6_rendering_npi;type:varchar(50)"     json:"service6_rendering_npi,omitempty"`
	FederalTaxID             *string    `gorm:"column:federal_tax_id;type:varchar(50)"             json:"federal_tax_id,omitempty"`
	TaxIDType                *string    `gorm:"column:tax_id_type;type:varchar(10)"                json:"tax_id_type,omitempty"`
	AcceptAssignment         *string    `gorm:"column:accept_assignment;type:varchar(10)"          json:"accept_assignment,omitempty"`
	PhysicianSignature       *string    `gorm:"column:physician_signature;type:varchar(255)"       json:"physician_signature,omitempty"`
	PhysicianSignatureDate   *string    `gorm:"column:physician_signature_date;type:varchar(50)"   json:"physician_signature_date,omitempty"`
	FacilityName             *string    `gorm:"column:facility_name;type:varchar(255)"             json:"facility_name,omitempty"`
	FacilityStreet           *string    `gorm:"column:facility_street;type:varchar(255)"           json:"facility_street,omitempty"`
	FacilityCityStateZip     *string    `gorm:"column:facility_city_state_zip;type:varchar(255)"   json:"facility_city_state_zip,omitempty"`
	FacilityNpi              *string    `gorm:"column:facility_npi;type:varchar(50)"               json:"facility_npi,omitempty"`
	BillingProviderName      *string    `gorm:"column:billing_provider_name;type:varchar(255)"     json:"billing_provider_name,omitempty"`
	BillingProviderStreet    *string    `gorm:"column:billing_provider_street;type:varchar(255)"   json:"billing_provider_street,omitempty"`
	BillingProviderCSZ       *string    `gorm:"column:billing_provider_city_state_zip;type:varchar(255)" json:"billing_provider_city_state_zip,omitempty"`
	BillingProviderPhoneArea *string    `gorm:"column:billing_provider_phone_area;type:varchar(10)" json:"billing_provider_phone_area,omitempty"`
	BillingProviderPhone     *string    `gorm:"column:billing_provider_phone;type:varchar(20)"     json:"billing_provider_phone,omitempty"`
	BillingProviderNpi       *string    `gorm:"column:billing_provider_npi;type:varchar(50)"       json:"billing_provider_npi,omitempty"`
	CreatedAt                *time.Time `gorm:"column:created_at;type:timestamptz;default:now()"   json:"created_at,omitempty"`
	UpdatedAt                *time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"   json:"updated_at,omitempty"`
}

func (ClaimTemplate) TableName() string { return "claim_template" }

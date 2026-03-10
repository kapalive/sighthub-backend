// internal/models/prescriptions/patient_prescription.go
package prescriptions

import "time"

// PatientPrescription ↔ table: patient_prescription
type PatientPrescription struct {
	IDPatientPrescription int64      `gorm:"column:id_patient_prescription;primaryKey;autoIncrement" json:"id_patient_prescription"`
	PatientID             int64      `gorm:"column:patient_id;not null"                              json:"patient_id"`
	PrescriptionDate      *time.Time `gorm:"column:prescription_date;type:date"                      json:"prescription_date,omitempty"`
	Note                  *string    `gorm:"column:note;type:text"                                   json:"note,omitempty"`
	Doctor                string     `gorm:"column:doctor;type:varchar(50);not null"                 json:"doctor"`
	NPI                   *string    `gorm:"column:npi;type:varchar(15);index"                       json:"npi,omitempty"`
	License               *string    `gorm:"column:license;type:varchar(50)"                         json:"license,omitempty"`
	LocationID            int        `gorm:"column:location_id;not null"                             json:"location_id"`
	Signature             *string    `gorm:"column:signature;type:text"                              json:"signature,omitempty"`
	Medication            *string    `gorm:"column:medication;type:text"                             json:"medication,omitempty"`
	Dosage                *string    `gorm:"column:dosage;type:text"                                 json:"dosage,omitempty"`
	DocumentLink          *string    `gorm:"column:document_link;type:varchar(255)"                  json:"document_link,omitempty"`
	DateUpload            *time.Time `gorm:"column:date_upload;type:timestamptz;autoCreateTime"      json:"date_upload,omitempty"`
	GOrC                  *string    `gorm:"column:g_or_c;type:varchar(50)"                          json:"g_or_c,omitempty"`
	PhoneNumber           *string    `gorm:"column:phone_number;type:varchar(15)"                    json:"phone_number,omitempty"`

	// preload
	GlassesPrescription     *GlassesPrescription     `gorm:"foreignKey:PrescriptionID;references:IDPatientPrescription" json:"-"`
	ContactLensPrescription *ContactLensPrescription `gorm:"foreignKey:PrescriptionID;references:IDPatientPrescription" json:"-"`
}

func (PatientPrescription) TableName() string { return "patient_prescription" }

func (p *PatientPrescription) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_patient_prescription": p.IDPatientPrescription,
		"patient_id":              p.PatientID,
		"note":                    p.Note,
		"doctor":                  p.Doctor,
		"npi":                     p.NPI,
		"license":                 p.License,
		"location_id":             p.LocationID,
		"signature":               p.Signature,
		"medication":              p.Medication,
		"dosage":                  p.Dosage,
		"document_link":           p.DocumentLink,
		"g_or_c":                  p.GOrC,
		"phone_number":            p.PhoneNumber,
	}
	if p.PrescriptionDate != nil {
		m["prescription_date"] = p.PrescriptionDate.Format("2006-01-02")
	} else {
		m["prescription_date"] = nil
	}
	if p.DateUpload != nil {
		m["date_upload"] = p.DateUpload.Format(time.RFC3339)
	} else {
		m["date_upload"] = nil
	}
	if p.GlassesPrescription != nil {
		m["glasses_prescription"] = p.GlassesPrescription.ToMap()
	}
	if p.ContactLensPrescription != nil {
		m["contact_lens_prescription"] = p.ContactLensPrescription.ToMap()
	}
	return m
}

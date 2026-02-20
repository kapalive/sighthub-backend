package patients

import (
	"time"
)

// Postgres enum types (по желанию оставь как string с type:...,
// если у тебя созданы ENUM gender/pronoun в БД).
type Gender string
type Pronoun string

const (
	GenderMale           Gender = "Male"
	GenderFemale         Gender = "Female"
	GenderOther          Gender = "Other"
	GenderPreferNotToSay Gender = "Prefer not to say"
	GenderUnknown        Gender = "Unknown"
)

// Patient ⇄ table: patient
type Patient struct {
	IDPatient  int64   `gorm:"column:id_patient;primaryKey"             json:"id_patient"`
	FirstName  string  `gorm:"column:first_name;type:varchar(50);not null" json:"first_name"`
	MiddleName *string `gorm:"column:middle_name;type:varchar(50)"    json:"middle_name,omitempty"`
	LastName   string  `gorm:"column:last_name;type:varchar(50);not null" json:"last_name"`

	DOB *time.Time `gorm:"column:dob;type:date" json:"-"`

	// ENUM gender (postgres type 'gender'); not null
	Gender Gender `gorm:"column:gender;type:gender;not null" json:"gender"`

	Phone     *string `gorm:"column:phone;type:varchar(15)"       json:"phone,omitempty"`
	PhoneHome *string `gorm:"column:phone_home;type:varchar(15)"  json:"phone_home,omitempty"`
	CellWork  *string `gorm:"column:cell_work;type:varchar(15)"   json:"cell_work,omitempty"`
	Email     *string `gorm:"column:email;type:varchar(100)"      json:"email,omitempty"`

	StreetAddress *string `gorm:"column:street_address;type:varchar(100)" json:"street_address,omitempty"`
	AddressLine2  *string `gorm:"column:address_line_2;type:varchar(100)" json:"address_line_2,omitempty"`
	City          *string `gorm:"column:city;type:varchar(50)"            json:"city,omitempty"`
	State         *string `gorm:"column:state;type:varchar(2)"            json:"state,omitempty"`
	ZipCode       *string `gorm:"column:zip_code;type:varchar(10)"        json:"zip_code,omitempty"`

	RaceID      *int64  `gorm:"column:race_id"            json:"race_id,omitempty"`
	EthnicityID *int64  `gorm:"column:ethnicity_id"       json:"ethnicity_id,omitempty"`
	SSN         *string `gorm:"column:ssn;type:varchar(11)"  json:"ssn,omitempty"`
	Pref        *string `gorm:"column:pref;type:varchar(100)" json:"pref,omitempty"`

	// pronoun/assigned_sex как enum в БД (если есть), иначе можно указать type:varchar(...)
	Pronoun     *string `gorm:"column:pronoun;type:pronoun" json:"pronoun,omitempty"`
	AssignedSex *string `gorm:"column:assigned_sex;type:gender" json:"assigned_sex,omitempty"`

	MailingList *bool `gorm:"column:mailing_list" json:"mailing_list,omitempty"`
	Survey      *bool `gorm:"column:survey"       json:"survey,omitempty"`

	PreferredLanguageID *int64 `gorm:"column:preferred_language_id" json:"preferred_language_id,omitempty"`

	// Связи можно добавить позже, когда будут пакеты:
	// PreferredLanguage *lang.PreferredLanguage `gorm:"foreignKey:PreferredLanguageID;references:IDPreferredLanguage" json:"preferred_language,omitempty"`
	// Race              *race.Race               `gorm:"foreignKey:RaceID;references:IDRace" json:"race,omitempty"`
	// Ethnicity         *ethnicity.Ethnicity     `gorm:"foreignKey:EthnicityID;references:IDEthnicity" json:"ethnicity,omitempty"`
}

func (Patient) TableName() string { return "patient" }

// ToMap — аналог Python to_dict()
func (p *Patient) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_patient":            p.IDPatient,
		"first_name":            p.FirstName,
		"middle_name":           p.MiddleName,
		"last_name":             p.LastName,
		"gender":                p.Gender,
		"phone":                 p.Phone,
		"phone_home":            p.PhoneHome,
		"cell_work":             p.CellWork,
		"email":                 p.Email,
		"street_address":        p.StreetAddress,
		"address_line_2":        p.AddressLine2,
		"city":                  p.City,
		"state":                 p.State,
		"zip_code":              p.ZipCode,
		"race_id":               p.RaceID,
		"ethnicity_id":          p.EthnicityID,
		"ssn":                   p.SSN,
		"pref":                  p.Pref,
		"pronoun":               p.Pronoun,
		"assigned_sex":          p.AssignedSex,
		"mailing_list":          p.MailingList,
		"survey":                p.Survey,
		"preferred_language_id": p.PreferredLanguageID,
	}
	if p.DOB != nil && !p.DOB.IsZero() {
		m["dob"] = p.DOB.Format("2006-01-02")
	} else {
		m["dob"] = nil
	}
	return m
}

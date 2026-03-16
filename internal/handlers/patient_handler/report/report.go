package report

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/patients"
)

type Handler struct{ DB *gorm.DB }

func New(db *gorm.DB) *Handler { return &Handler{DB: db} }

type patientRow struct {
	patients.Patient
	RaceName          *string `gorm:"column:race_name"`
	EthnicityName     *string `gorm:"column:ethnicity_name"`
	LocationName      *string `gorm:"column:location"`
	PreferredLanguage *string `gorm:"column:preferred_language"`
}

// GET /api/patient/report/all_patients/csv
func (h *Handler) AllPatientsCSV(w http.ResponseWriter, r *http.Request) {
	var rows []patientRow
	h.DB.Model(&patients.Patient{}).
		Select(`patient.*,
			race.race_name AS race_name,
			ethnicity.ethnicity_name AS ethnicity_name,
			location.full_name AS "location",
			preferred_language.language AS preferred_language`).
		Joins("LEFT JOIN race ON patient.race_id = race.id_race").
		Joins("LEFT JOIN ethnicity ON patient.ethnicity_id = ethnicity.id_ethnicity").
		Joins("LEFT JOIN location ON patient.location_id = location.id_location").
		Joins("LEFT JOIN preferred_language ON patient.preferred_language_id = preferred_language.id_preferred_language").
		Scan(&rows)

	w.Header().Set("Content-Disposition", "attachment; filename=patients.csv")
	w.Header().Set("Content-Type", "text/csv")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	headers := []string{
		"id_patient", "first_name", "middle_name", "last_name", "dob", "gender",
		"phone", "phone_home", "cell_work", "email", "street_address",
		"address_line_2", "city", "state", "zip_code", "race_name",
		"ethnicity_name", "ssn", "pref", "pronoun", "assigned_sex",
		"mailing_list", "location", "survey", "preferred_language",
	}
	cw.Write(headers) //nolint:errcheck

	for _, p := range rows {
		cw.Write([]string{
			fmt.Sprintf("%d", p.IDPatient),
			p.FirstName,
			derefStr(p.MiddleName),
			p.LastName,
			fmtTime(p.DOB),
			fmtGender(p.Gender),
			derefStr(p.Phone),
			derefStr(p.PhoneHome),
			derefStr(p.CellWork),
			derefStr(p.Email),
			derefStr(p.StreetAddress),
			derefStr(p.AddressLine2),
			derefStr(p.City),
			derefStr(p.State),
			derefStr(p.ZipCode),
			derefStr(p.RaceName),
			derefStr(p.EthnicityName),
			derefStr(p.SSN),
			derefStr(p.Pref),
			derefStr(p.Pronoun),
			derefStr(p.AssignedSex),
			fmtBool(p.MailingList),
			derefStr(p.LocationName),
			fmtBool(p.Survey),
			derefStr(p.PreferredLanguage),
		}) //nolint:errcheck
	}

}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func fmtTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func fmtBool(b *bool) string {
	if b == nil {
		return ""
	}
	if *b {
		return "true"
	}
	return "false"
}

func fmtGender(g *patients.Gender) string {
	if g == nil {
		return ""
	}
	return string(*g)
}

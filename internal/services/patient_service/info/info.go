package info

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	apptModel "sighthub-backend/internal/models/appointment"
	empModel "sighthub-backend/internal/models/employees"
	locModel "sighthub-backend/internal/models/location"
	patModel "sighthub-backend/internal/models/patients"
	presModel "sighthub-backend/internal/models/prescriptions"
	pkgActivity "sighthub-backend/pkg/activitylog"
)

// ─── Service ─────────────────────────────────────────────────────────────────

type Service struct {
	db     *gorm.DB
	dbName string
}

func New(db *gorm.DB, dbName string) *Service {
	return &Service{db: db, dbName: dbName}
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

type PatientDetail struct {
	ID                int64       `json:"id"`
	FirstName         string      `json:"first_name"`
	MiddleName        string      `json:"middle_name"`
	LastName          string      `json:"last_name"`
	DOB               string      `json:"dob"`
	Gender            string      `json:"gender"`
	Phone             string      `json:"phone"`
	PhoneHome         string      `json:"phone_home"`
	CellWork          string      `json:"cell_work"`
	Email             string      `json:"email"`
	StreetAddress     string      `json:"street_address"`
	AddressLine2      string      `json:"address_line_2"`
	City              string      `json:"city"`
	State             string      `json:"state"`
	ZipCode           string      `json:"zip_code"`
	SSN               string      `json:"ssn"`
	Pref              string      `json:"pref"`
	Pronoun           string      `json:"pronoun"`
	AssignedSex       string      `json:"assigned_sex"`
	MailingList       bool        `json:"mailing_list"`
	Survey            bool        `json:"survey"`
	PreferredLanguage interface{} `json:"preferred_language"`
	PatientCredit     string      `json:"patient_credit"`
}

type PatientSearchItem struct {
	ID        int64  `json:"id"`
	LastName  string `json:"last_name"`
	FirstName string `json:"first_name"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	DOB       string `json:"dob,omitempty"`
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, nil, err
	}

	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, nil, err
	}

	if emp.LocationID == nil {
		return &emp, nil, nil
	}

	var loc locModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return &emp, nil, err
	}

	return &emp, &loc, nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func ptrInt64(v int64) *int64 { return &v }

// ─── Languages ────────────────────────────────────────────────────────────────

func (s *Service) GetLanguages() ([]map[string]interface{}, error) {
	var langs []patModel.PreferredLanguage
	if err := s.db.Find(&langs).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(langs))
	for i, l := range langs {
		result[i] = map[string]interface{}{
			"id_preferred_language": l.IDPreferredLanguage,
			"lang":                  l.Language,
		}
	}
	return result, nil
}

// ─── Country codes ────────────────────────────────────────────────────────────

func (s *Service) GetCountryCodes() ([]map[string]interface{}, error) {
	type row struct {
		Country   string  `gorm:"column:country"`
		PhoneCode *string `gorm:"column:phone_code"`
	}
	var rows []row
	if err := s.db.Raw("SELECT country, phone_code FROM country WHERE phone_code IS NOT NULL ORDER BY country ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		result[i] = map[string]interface{}{
			"code":    r.PhoneCode,
			"country": r.Country,
		}
	}
	return result, nil
}

// ─── Create patient ───────────────────────────────────────────────────────────

type CreatePatientInput struct {
	FirstName           string  `json:"first_name"`
	MiddleName          *string `json:"middle_name"`
	LastName            string  `json:"last_name"`
	DOB                 *string `json:"dob"`
	Gender              string  `json:"gender"`
	Phone               *string `json:"phone"`
	PhoneHome           *string `json:"phone_home"`
	CellWork            *string `json:"cell_work"`
	Email               *string `json:"email"`
	StreetAddress       *string `json:"street_address"`
	AddressLine2        *string `json:"address_line_2"`
	City                *string `json:"city"`
	State               *string `json:"state"`
	ZipCode             *string `json:"zip_code"`
	SSN                 *string `json:"ssn"`
	Pref                *string `json:"pref"`
	Pronoun             *string `json:"pronoun"`
	AssignedSex         *string `json:"assigned_sex"`
	MailingList         *bool   `json:"mailing_list"`
	Survey              *bool   `json:"survey"`
	PreferredLanguageID *int64  `json:"preferred_language_id"`
}

func (s *Service) CreatePatient(username string, input CreatePatientInput) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil || loc == nil {
		return nil, errors.New("employee or location not found")
	}

	if input.FirstName == "" || input.LastName == "" {
		return nil, errors.New("first_name and last_name are required")
	}

	if input.PreferredLanguageID != nil {
		var lang patModel.PreferredLanguage
		if err := s.db.First(&lang, *input.PreferredLanguageID).Error; err != nil {
			return nil, errors.New("invalid preferred_language_id")
		}
	}

	patient := patModel.Patient{
		LocationID:          ptrInt64(int64(loc.IDLocation)),
		FirstName:           input.FirstName,
		MiddleName:          input.MiddleName,
		LastName:            input.LastName,
		Gender:              patModel.Gender(input.Gender),
		Phone:               input.Phone,
		PhoneHome:           input.PhoneHome,
		CellWork:            input.CellWork,
		Email:               input.Email,
		StreetAddress:       input.StreetAddress,
		AddressLine2:        input.AddressLine2,
		City:                input.City,
		State:               input.State,
		ZipCode:             input.ZipCode,
		SSN:                 input.SSN,
		Pref:                input.Pref,
		Pronoun:             input.Pronoun,
		AssignedSex:         input.AssignedSex,
		MailingList:         input.MailingList,
		Survey:              input.Survey,
		PreferredLanguageID: input.PreferredLanguageID,
	}

	if input.DOB != nil {
		t, err := time.Parse("2006-01-02", *input.DOB)
		if err != nil {
			return nil, errors.New("dob must be YYYY-MM-DD")
		}
		patient.DOB = &t
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&patient).Error; err != nil {
			return err
		}
		return pkgActivity.Log(tx, "patient", "create",
			pkgActivity.WithEntity(patient.IDPatient),
			pkgActivity.WithDetails(map[string]interface{}{
				"first_name": patient.FirstName,
				"last_name":  patient.LastName,
			}),
		)
	}); err != nil {
		return nil, err
	}

	return patient.ToMap(), nil
}

// ─── Get patient ──────────────────────────────────────────────────────────────

func (s *Service) GetPatient(username string, patientID int64) (*PatientDetail, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil || loc == nil {
		return nil, errors.New("employee or location not found")
	}

	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	if patient.LocationID == nil || *patient.LocationID != int64(loc.IDLocation) {
		return nil, errors.New("forbidden")
	}

	// credit
	var credit float64
	var cb patModel.ClientBalance
	if err := s.db.Where("patient_id = ? AND location_id = ?", patientID, loc.IDLocation).First(&cb).Error; err == nil {
		credit = cb.Credit
	}

	// recently viewed upsert
	var rvp patModel.RecentlyViewedPatient
	if err := s.db.Where("location_id = ? AND patient_id = ?", loc.IDLocation, patientID).First(&rvp).Error; err == nil {
		s.db.Model(&rvp).Update("datetime_viewed", gorm.Expr("NOW()"))
	} else {
		s.db.Create(&patModel.RecentlyViewedPatient{
			LocationID: loc.IDLocation,
			PatientID:  patientID,
		})
	}

	// preferred_language
	var prefLang interface{}
	if patient.PreferredLanguageID != nil {
		var lang patModel.PreferredLanguage
		if s.db.First(&lang, *patient.PreferredLanguageID).Error == nil {
			prefLang = map[string]interface{}{
				"id_preferred_language": lang.IDPreferredLanguage,
				"lang":                  lang.Language,
			}
		}
	}

	dob := ""
	if patient.DOB != nil {
		dob = patient.DOB.Format("2006-01-02")
	}

	return &PatientDetail{
		ID:                patient.IDPatient,
		FirstName:         patient.FirstName,
		MiddleName:        derefStr(patient.MiddleName),
		LastName:          patient.LastName,
		DOB:               dob,
		Gender:            string(patient.Gender),
		Phone:             derefStr(patient.Phone),
		PhoneHome:         derefStr(patient.PhoneHome),
		CellWork:          derefStr(patient.CellWork),
		Email:             derefStr(patient.Email),
		StreetAddress:     derefStr(patient.StreetAddress),
		AddressLine2:      derefStr(patient.AddressLine2),
		City:              derefStr(patient.City),
		State:             derefStr(patient.State),
		ZipCode:           derefStr(patient.ZipCode),
		SSN:               derefStr(patient.SSN),
		Pref:              derefStr(patient.Pref),
		Pronoun:           derefStr(patient.Pronoun),
		AssignedSex:       derefStr(patient.AssignedSex),
		MailingList:       derefBool(patient.MailingList),
		Survey:            derefBool(patient.Survey),
		PreferredLanguage: prefLang,
		PatientCredit:     fmt.Sprintf("%.2f", credit),
	}, nil
}

// ─── Update patient ───────────────────────────────────────────────────────────

type UpdatePatientInput struct {
	FirstName           *string `json:"first_name"`
	MiddleName          *string `json:"middle_name"`
	LastName            *string `json:"last_name"`
	DOB                 *string `json:"dob"`
	Gender              *string `json:"gender"`
	Phone               *string `json:"phone"`
	PhoneHome           *string `json:"phone_home"`
	CellWork            *string `json:"cell_work"`
	Email               *string `json:"email"`
	StreetAddress       *string `json:"street_address"`
	AddressLine2        *string `json:"address_line_2"`
	City                *string `json:"city"`
	State               *string `json:"state"`
	ZipCode             *string `json:"zip_code"`
	SSN                 *string `json:"ssn"`
	Pref                *string `json:"pref"`
	Pronoun             *string `json:"pronoun"`
	AssignedSex         *string `json:"assigned_sex"`
	MailingList         *bool   `json:"mailing_list"`
	Survey              *bool   `json:"survey"`
	PreferredLanguageID *int64  `json:"preferred_language_id"`
}

func (s *Service) UpdatePatient(patientID int64, input UpdatePatientInput) error {
	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return errors.New("patient not found")
	}

	updates := map[string]interface{}{}

	if input.FirstName != nil       { updates["first_name"] = *input.FirstName }
	if input.MiddleName != nil      { updates["middle_name"] = *input.MiddleName }
	if input.LastName != nil        { updates["last_name"] = *input.LastName }
	if input.Gender != nil          { updates["gender"] = *input.Gender }
	if input.Phone != nil           { updates["phone"] = *input.Phone }
	if input.PhoneHome != nil       { updates["phone_home"] = *input.PhoneHome }
	if input.CellWork != nil        { updates["cell_work"] = *input.CellWork }
	if input.Email != nil           { updates["email"] = *input.Email }
	if input.StreetAddress != nil   { updates["street_address"] = *input.StreetAddress }
	if input.AddressLine2 != nil    { updates["address_line_2"] = *input.AddressLine2 }
	if input.City != nil            { updates["city"] = *input.City }
	if input.State != nil           { updates["state"] = *input.State }
	if input.ZipCode != nil         { updates["zip_code"] = *input.ZipCode }
	if input.SSN != nil             { updates["ssn"] = *input.SSN }
	if input.Pref != nil            { updates["pref"] = *input.Pref }
	if input.Pronoun != nil         { updates["pronoun"] = *input.Pronoun }
	if input.AssignedSex != nil     { updates["assigned_sex"] = *input.AssignedSex }
	if input.MailingList != nil     { updates["mailing_list"] = *input.MailingList }
	if input.Survey != nil          { updates["survey"] = *input.Survey }
	if input.PreferredLanguageID != nil { updates["preferred_language_id"] = *input.PreferredLanguageID }

	if input.DOB != nil {
		t, err := time.Parse("2006-01-02", *input.DOB)
		if err != nil {
			return errors.New("dob must be YYYY-MM-DD")
		}
		updates["dob"] = t
	}

	if len(updates) == 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&patient).Updates(updates).Error; err != nil {
			return err
		}
		return pkgActivity.Log(tx, "patient", "update",
			pkgActivity.WithEntity(patientID),
			pkgActivity.WithDetails(map[string]interface{}{
				"first_name": patient.FirstName,
				"last_name":  patient.LastName,
			}),
		)
	})
}

// ─── Delete patient ───────────────────────────────────────────────────────────

func (s *Service) DeletePatient(patientID int64) error {
	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return errors.New("patient not found")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := pkgActivity.Log(tx, "patient", "delete",
			pkgActivity.WithEntity(patientID),
			pkgActivity.WithDetails(map[string]interface{}{
				"first_name": patient.FirstName,
				"last_name":  patient.LastName,
			}),
		); err != nil {
			return err
		}
		return tx.Delete(&patient).Error
	})
}

// ─── Generate filenames ───────────────────────────────────────────────────────

func (s *Service) GenerateFilenameDoc(patientID int64) (map[string]interface{}, error) {
	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}
	var nextID int64
	if err := s.db.Raw("SELECT nextval('documents_patient_id_documents_patient_seq')").Scan(&nextID).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"filename":    fmt.Sprintf("documents_patient-%d_%s.pdf", nextID, s.dbName),
		"reserved_id": nextID,
	}, nil
}

func (s *Service) GenerateFilenamePrescription(patientID int64) (map[string]interface{}, error) {
	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}
	var nextID int64
	if err := s.db.Raw("SELECT nextval('patient_prescription_id_patient_prescription_seq')").Scan(&nextID).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"filename": fmt.Sprintf("patient_prescription-%d_%s.pdf", nextID, s.dbName),
	}, nil
}

func (s *Service) GenerateFilenameInsurancePolicy(patientID int64) (map[string]interface{}, error) {
	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}
	var nextID int64
	if err := s.db.Raw("SELECT nextval('insurance_policy_id_insurance_policy_seq')").Scan(&nextID).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"filename": fmt.Sprintf("insurance_policy-%d_%s.pdf", nextID, s.dbName),
	}, nil
}

// ─── Search patients ──────────────────────────────────────────────────────────

type SearchParams struct {
	FirstName *string
	LastName  *string
	DOB       *string
	City      *string
	State     *string
	Phone     *string
	Email     *string
}

func (s *Service) SearchPatients(username string, params SearchParams) ([]PatientSearchItem, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil || loc == nil {
		return nil, errors.New("employee or location not found")
	}

	query := s.db.Model(&patModel.Patient{}).Where("location_id = ?", loc.IDLocation)

	if params.FirstName != nil { query = query.Where("first_name ILIKE ?", *params.FirstName+"%") }
	if params.LastName != nil  { query = query.Where("last_name ILIKE ?", *params.LastName+"%") }
	if params.City != nil      { query = query.Where("city ILIKE ?", *params.City+"%") }
	if params.State != nil     { query = query.Where("state ILIKE ?", *params.State+"%") }
	if params.Phone != nil     { query = query.Where("phone ILIKE ?", *params.Phone+"%") }
	if params.Email != nil     { query = query.Where("email ILIKE ?", *params.Email+"%") }

	if params.DOB != nil {
		t, err := time.Parse("2006-01-02", *params.DOB)
		if err != nil {
			return nil, errors.New("dob must be YYYY-MM-DD")
		}
		query = query.Where("dob = ?", t)
	}

	var patients []patModel.Patient
	if err := query.Order("last_name ASC, first_name ASC").Limit(25).Find(&patients).Error; err != nil {
		return nil, err
	}

	result := make([]PatientSearchItem, len(patients))
	for i, p := range patients {
		var addrParts []string
		for _, part := range []*string{p.StreetAddress, p.City, p.State, p.ZipCode} {
			if part != nil && *part != "" {
				addrParts = append(addrParts, *part)
			}
		}

		dob := ""
		if p.DOB != nil {
			dob = p.DOB.Format("2006-01-02")
		}

		result[i] = PatientSearchItem{
			ID:        p.IDPatient,
			LastName:  p.LastName,
			FirstName: p.FirstName,
			Address:   strings.Join(addrParts, ", "),
			Phone:     derefStr(p.Phone),
			DOB:       dob,
		}
	}
	return result, nil
}

// ─── Patient prescriptions (uploaded PDF list) ────────────────────────────────

func (s *Service) GetPatientPrescriptions(patientID int64) ([]map[string]interface{}, error) {
	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	var prescriptions []presModel.PatientPrescription
	if err := s.db.Where("patient_id = ?", patientID).Find(&prescriptions).Error; err != nil {
		return nil, err
	}

	downloadURL := os.Getenv("DOWNLOAD_API_URL")
	if downloadURL == "" {
		downloadURL = "http://172.16.6.15:8080/download"
	}

	result := make([]map[string]interface{}, 0, len(prescriptions))
	for _, p := range prescriptions {
		if p.PrescriptionDate == nil {
			continue
		}
		filePath := fmt.Sprintf("patient_prescription-%d.pdf", p.IDPatientPrescription)
		result = append(result, map[string]interface{}{
			"date":          p.PrescriptionDate.Format("2006-01-02"),
			"description":   p.Note,
			"download_path": downloadURL + "/" + filePath,
		})
	}
	return result, nil
}

// ─── Patient appointments ─────────────────────────────────────────────────────

func (s *Service) GetPatientAppointments(patientID int64) ([]map[string]interface{}, error) {
	var appts []apptModel.Appointment
	if err := s.db.
		Preload("Schedule.Employee").
		Preload("Reason").
		Where("patient_id = ?", patientID).
		Order("appointment_date DESC").
		Find(&appts).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(appts))
	for i, a := range appts {
		doctorName := "Unknown Doctor"
		if a.Schedule != nil && a.Schedule.Employee != nil {
			doctorName = fmt.Sprintf("Dr. %s %s", a.Schedule.Employee.FirstName, a.Schedule.Employee.LastName)
		}

		title := "No reason provided"
		if a.Reason != nil {
			title = a.Reason.Reason
		}

		result[i] = map[string]interface{}{
			"date":   a.AppointmentDate.Format("02/01/2006"),
			"title":  title,
			"doctor": doctorName,
		}
	}
	return result, nil
}

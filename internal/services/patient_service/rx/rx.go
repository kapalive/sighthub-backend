package rx

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	locModel "sighthub-backend/internal/models/location"
	patModel "sighthub-backend/internal/models/patients"
	presModel "sighthub-backend/internal/models/prescriptions"
)

// ─── Service ──────────────────────────────────────────────────────────────────

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── Input DTOs ───────────────────────────────────────────────────────────────

type GlassesInput struct {
	OdSph             *string      `json:"od_sph"`
	OsSph             *string      `json:"os_sph"`
	OdCyl             *string      `json:"od_cyl"`
	OsCyl             *string      `json:"os_cyl"`
	OdAxis            *string      `json:"od_axis"`
	OsAxis            *string      `json:"os_axis"`
	OdAdd             *FlexFloat `json:"od_add"`
	OsAdd             *FlexFloat `json:"os_add"`
	OdHPrism          *FlexFloat `json:"od_h_prism"`
	OsHPrism          *FlexFloat `json:"os_h_prism"`
	OdHPrismDirection *string      `json:"od_h_prism_direction"`
	OsHPrismDirection *string      `json:"os_h_prism_direction"`
	OdVPrism          *FlexFloat `json:"od_v_prism"`
	OsVPrism          *FlexFloat `json:"os_v_prism"`
	OdVPrismDirection *string      `json:"od_v_prism_direction"`
	OsVPrismDirection *string      `json:"os_v_prism_direction"`
	OdDpd             *FlexFloat `json:"od_dpd"`
	OsDpd             *FlexFloat `json:"os_dpd"`
	Date              *string      `json:"date"`            // expiration_date alias from frontend
	ExpirationDate    *string      `json:"expiration_date"` // direct
}

type ContactsInput struct {
	OdContLens     *string      `json:"od_cont_lens"`
	OsContLens     *string      `json:"os_cont_lens"`
	OdBc           *string      `json:"od_bc"`
	OsBc           *string      `json:"os_bc"`
	OdDia          *FlexFloat `json:"od_dia"`
	OsDia          *FlexFloat `json:"os_dia"`
	OdPwr          *string      `json:"od_pwr"`
	OsPwr          *string      `json:"os_pwr"`
	OdCyl          *string      `json:"od_cyl"`
	OsCyl          *string      `json:"os_cyl"`
	OdAxis         *string      `json:"od_axis"`
	OsAxis         *string      `json:"os_axis"`
	OdAdd          *string      `json:"od_add"`
	OsAdd          *string      `json:"os_add"`
	OdColor        *string      `json:"od_color"`
	OsColor        *string      `json:"os_color"`
	OdType         *string      `json:"od_type"`
	OsType         *string      `json:"os_type"`
	Date           *string      `json:"date"`            // expiration_date alias from frontend
	ExpirationDate *string      `json:"expiration_date"` // direct
}

type PrescriptionPayload struct {
	GlassesData  *GlassesInput  `json:"glasses_data"`
	ContactsData *ContactsInput `json:"contacts_data"`
}

type CreateRxInput struct {
	PatientID      int64               `json:"patient_id"`
	Doctor         *string             `json:"doctor"`
	DoctorPhone    *string             `json:"doctor_phone"`
	NPI            *string             `json:"npi"`
	License        *string             `json:"license"`
	PrescriptionDate *string           `json:"prescription_date"`
	Signature      *string             `json:"signature"`
	Note           *string             `json:"note"`
	Medication     *string             `json:"medication"`
	Dosage         *string             `json:"dosage"`
	DocumentLink   *string             `json:"document_link"`
	Prescription   *PrescriptionPayload `json:"prescription"`
}

type UpdateRxInput struct {
	IDRx         int64               `json:"id_rx"`
	Date         *string             `json:"date"` // prescription_date
	Note         *string             `json:"note"`
	Medication   *string             `json:"medication"`
	Dosage       *string             `json:"dosage"`
	DocumentLink *string             `json:"document_link"`
	Doctor       *string             `json:"doctor"`
	DoctorPhone  *string             `json:"doctor_phone"`
	NPI          *string             `json:"npi"`
	License      *string             `json:"license"`
	Signature    *string             `json:"signature"`
	LocationID   *int                `json:"location_id"`
	Prescription *PrescriptionPayload `json:"prescription"`
}

// ─── GET /rx/amidoctor ────────────────────────────────────────────────────────

func (s *Service) AmIDoctor(username string) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil || emp == nil {
		return nil, errors.New("employee not found")
	}

	// job_title_id = 1 means doctor
	if emp.JobTitleID == nil || *emp.JobTitleID != int64(1) {
		return nil, nil // not a doctor
	}

	var locationName *string
	if loc != nil {
		locationName = &loc.FullName
	}

	result := map[string]interface{}{
		"doctor":   "Dr. " + emp.FirstName + " " + emp.LastName,
		"npi":      nil,
		"ein":      nil,
		"location": locationName,
	}

	// Look up NPI/EIN from doctor_npi_number
	type npiRow struct {
		NPI *string `gorm:"column:dr_npi_number"`
		EIN *string `gorm:"column:ein"`
	}
	var npi npiRow
	if s.db.Table("doctor_npi_number").
		Where("employee_id = ?", emp.IDEmployee).
		First(&npi).Error == nil {
		result["npi"] = npi.NPI
		result["ein"] = npi.EIN
	}

	return result, nil
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

func emptyStr(s *string) bool {
	return s == nil || *s == "" || *s == " "
}

func hasGlassesData(g *presModel.GlassesPrescription) bool {
	if g == nil {
		return false
	}
	return g.OdSph != nil || g.OsSph != nil || g.OdCyl != nil || g.OsCyl != nil ||
		g.OdAxis != nil || g.OsAxis != nil || g.OdAdd != nil || g.OsAdd != nil
}

func hasContactsData(c *presModel.ContactLensPrescription) bool {
	if c == nil {
		return false
	}
	return c.OdContLens != nil || c.OsContLens != nil || c.OdBc != nil || c.OsBc != nil ||
		c.OdDia != nil || c.OsDia != nil || c.OdPwr != nil || c.OsPwr != nil
}

func computeGOrC(g *presModel.GlassesPrescription, c *presModel.ContactLensPrescription) *string {
	hasG := hasGlassesData(g)
	hasC := hasContactsData(c)
	if hasG && hasC {
		v := "G, C"
		return &v
	} else if hasG {
		v := "G"
		return &v
	} else if hasC {
		v := "C"
		return &v
	}
	return nil
}

func parseExpDate(dateStr *string, altStr *string) (*time.Time, error) {
	raw := ""
	if dateStr != nil && *dateStr != "" && *dateStr != " " {
		raw = *dateStr
	} else if altStr != nil && *altStr != "" && *altStr != " " {
		raw = *altStr
	}
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, errors.New("invalid date, expected YYYY-MM-DD")
	}
	return &t, nil
}

// FlexFloat accepts both JSON number (1.50) and JSON string ("1.50")
type FlexFloat struct {
	Value *float64
}

func (f *FlexFloat) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		f.Value = nil
		return nil
	}
	s := strings.Trim(string(b), `"`)
	if s == "" {
		f.Value = nil
		return nil
	}
	var v float64
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return err
	}
	f.Value = &v
	return nil
}

func ffVal(f *FlexFloat) *float64 {
	if f == nil {
		return nil
	}
	return f.Value
}

func normalizeEnumStr(s *string) *string {
	if s == nil || *s == "" || *s == " " {
		return nil
	}
	return s
}

func locationShortName(loc *locModel.Location) string {
	if loc == nil {
		return "Unknown location"
	}
	if loc.ShortName != nil {
		return *loc.ShortName
	}
	return "Unknown location"
}

func locationFullName(loc *locModel.Location) string {
	if loc == nil {
		return "Unknown Location"
	}
	return loc.FullName
}

func glassesToMap(g *presModel.GlassesPrescription) map[string]interface{} {
	if g == nil {
		return map[string]interface{}{}
	}
	m := g.ToMap()
	delete(m, "id_glasses_prescription")
	delete(m, "prescription_id")
	return m
}

func contactsToMap(c *presModel.ContactLensPrescription) map[string]interface{} {
	if c == nil {
		return map[string]interface{}{}
	}
	m := c.ToMap()
	delete(m, "id_contact_lens_prescription")
	delete(m, "prescription_id")
	return m
}

// ─── GET /latest-rx ───────────────────────────────────────────────────────────

func (s *Service) GetLatestRx(patientID int64) (map[string]interface{}, error) {
	var prescriptions []presModel.PatientPrescription
	s.db.Where("patient_id = ?", patientID).
		Order("prescription_date DESC").
		Find(&prescriptions)

	if len(prescriptions) == 0 {
		return map[string]interface{}{}, nil
	}

	// Priority: first with "G", then with "C"
	var prescription *presModel.PatientPrescription
	for i := range prescriptions {
		if strings.Contains(derefStr(prescriptions[i].GOrC), "G") {
			prescription = &prescriptions[i]
			break
		}
	}
	if prescription == nil {
		for i := range prescriptions {
			if derefStr(prescriptions[i].GOrC) == "C" {
				prescription = &prescriptions[i]
				break
			}
		}
	}
	if prescription == nil {
		return map[string]interface{}{}, nil
	}

	dateStr := ""
	if prescription.PrescriptionDate != nil {
		dateStr = prescription.PrescriptionDate.Format("2006-01-02")
	}

	resp := map[string]interface{}{
		"date":  dateStr,
		"id_rx": prescription.IDPatientPrescription,
	}

	gOrC := derefStr(prescription.GOrC)

	if strings.Contains(gOrC, "G") {
		var g presModel.GlassesPrescription
		if s.db.Where("prescription_id = ?", prescription.IDPatientPrescription).First(&g).Error == nil {
			gd := map[string]interface{}{
				"od_sph":    derefStr(g.OdSph),
				"os_sph":    derefStr(g.OsSph),
				"od_cyl":    derefStr(g.OdCyl),
				"os_cyl":    derefStr(g.OsCyl),
				"od_axis":   derefStr(g.OdAxis),
				"os_axis":   derefStr(g.OsAxis),
				"od_add":    g.OdAdd,
				"os_add":    g.OsAdd,
				"od_h_prism": g.OdHPrism,
				"os_h_prism": g.OsHPrism,
				"od_v_prism": g.OdVPrism,
				"os_v_prism": g.OsVPrism,
			}
			if g.ExpirationDate != nil {
				gd["expiration_date"] = g.ExpirationDate.Format("2006-01-02")
			} else {
				gd["expiration_date"] = nil
			}
			resp["glasses_data"] = gd
		}
	} else if gOrC == "C" {
		var c presModel.ContactLensPrescription
		if s.db.Where("prescription_id = ?", prescription.IDPatientPrescription).First(&c).Error == nil {
			cd := map[string]interface{}{
				"od_bc":   derefStr(c.OdBc),
				"os_bc":   derefStr(c.OsBc),
				"od_dia":  c.OdDia,
				"os_dia":  c.OsDia,
				"od_pwr":  derefStr(c.OdPwr),
				"os_pwr":  derefStr(c.OsPwr),
				"od_cyl":  derefStr(c.OdCyl),
				"os_cyl":  derefStr(c.OsCyl),
				"od_axis": derefStr(c.OdAxis),
				"os_axis": derefStr(c.OsAxis),
				"od_add":  derefStr(c.OdAdd),
				"os_add":  derefStr(c.OsAdd),
			}
			if c.ExpirationDate != nil {
				cd["expiration_date"] = c.ExpirationDate.Format("2006-01-02")
			} else {
				cd["expiration_date"] = nil
			}
			resp["contacts_data"] = cd
		}
	}

	return resp, nil
}

// ─── GET /rx-list ─────────────────────────────────────────────────────────────

func (s *Service) GetRxList(patientID int64) ([]map[string]interface{}, error) {
	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	var prescriptions []presModel.PatientPrescription
	s.db.Where("patient_id = ?", patientID).Find(&prescriptions)

	result := make([]map[string]interface{}, 0, len(prescriptions))
	for _, p := range prescriptions {
		locShort := "Unknown location"
		locFull := "Unknown Location"

		var loc locModel.Location
		if s.db.First(&loc, p.LocationID).Error == nil {
			locShort = locationShortName(&loc)
			locFull = locationFullName(&loc)
		}

		var date interface{}
		if p.PrescriptionDate != nil {
			date = p.PrescriptionDate.Format("2006-01-02")
		}

		var expDate interface{}
		gOrC := derefStr(p.GOrC)
		if strings.Contains(gOrC, "G") {
			var g presModel.GlassesPrescription
			if s.db.Where("prescription_id = ?", p.IDPatientPrescription).First(&g).Error == nil && g.ExpirationDate != nil {
				expDate = g.ExpirationDate.Format("2006-01-02")
			}
		} else if gOrC == "C" {
			var c presModel.ContactLensPrescription
			if s.db.Where("prescription_id = ?", p.IDPatientPrescription).First(&c).Error == nil && c.ExpirationDate != nil {
				expDate = c.ExpirationDate.Format("2006-01-02")
			}
		}

		result = append(result, map[string]interface{}{
			"id_rx":           p.IDPatientPrescription,
			"date":            date,
			"g_or_c":          gOrC,
			"location":        locShort,
			"doctor":          p.Doctor,
			"type_location":   locFull,
			"expiration_date": expDate,
		})
	}
	return result, nil
}

// ─── GET /rx ──────────────────────────────────────────────────────────────────

func (s *Service) GetRx(rxID int64) (map[string]interface{}, error) {
	var p presModel.PatientPrescription
	if err := s.db.First(&p, rxID).Error; err != nil {
		return nil, errors.New("prescription not found")
	}

	locShort := "Unknown location"
	locFull := "Unknown Location"
	var loc locModel.Location
	if s.db.First(&loc, p.LocationID).Error == nil {
		locShort = locationShortName(&loc)
		locFull = locationFullName(&loc)
	}

	var glasses presModel.GlassesPrescription
	gMap := map[string]interface{}{}
	if s.db.Where("prescription_id = ?", rxID).First(&glasses).Error == nil {
		gMap = glassesToMap(&glasses)
	}

	var contacts presModel.ContactLensPrescription
	cMap := map[string]interface{}{}
	if s.db.Where("prescription_id = ?", rxID).First(&contacts).Error == nil {
		cMap = contactsToMap(&contacts)
	}

	var date interface{}
	if p.PrescriptionDate != nil {
		date = p.PrescriptionDate.Format("2006-01-02")
	}

	return map[string]interface{}{
		"id_rx":        p.IDPatientPrescription,
		"signature":    p.Signature,
		"date":         date,
		"g_or_c":       derefStr(p.GOrC),
		"location":     locShort,
		"doctor":       p.Doctor,
		"doctor_phone": p.PhoneNumber,
		"npi":          p.NPI,
		"license":      p.License,
		"type_location": locFull,
		"note":         p.Note,
		"prescription": map[string]interface{}{
			"glasses_data":  gMap,
			"contacts_data": cMap,
		},
	}, nil
}

// ─── POST /rx ─────────────────────────────────────────────────────────────────

func (s *Service) CreateRx(username string, input CreateRxInput) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil || loc == nil {
		return nil, errors.New("employee or location not found")
	}

	var patient patModel.Patient
	if err := s.db.First(&patient, input.PatientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	// prescription_date
	var prescDate time.Time
	if input.PrescriptionDate != nil && *input.PrescriptionDate != "" && *input.PrescriptionDate != " " {
		t, err := time.Parse("2006-01-02", *input.PrescriptionDate)
		if err != nil {
			return nil, errors.New("invalid prescription_date, expected YYYY-MM-DD")
		}
		prescDate = t
	} else {
		prescDate = time.Now().Truncate(24 * time.Hour)
	}

	pres := presModel.PatientPrescription{
		PatientID:        input.PatientID,
		PrescriptionDate: &prescDate,
		Doctor:           derefStr(input.Doctor),
		PhoneNumber:      input.DoctorPhone,
		NPI:              input.NPI,
		License:          input.License,
		LocationID:       loc.IDLocation,
		Note:             input.Note,
		Signature:        input.Signature,
		Medication:       input.Medication,
		Dosage:           input.Dosage,
		DocumentLink:     input.DocumentLink,
	}

	var gInput GlassesInput
	var cInput ContactsInput
	if input.Prescription != nil {
		if input.Prescription.GlassesData != nil {
			gInput = *input.Prescription.GlassesData
		}
		if input.Prescription.ContactsData != nil {
			cInput = *input.Prescription.ContactsData
		}
	}

	// normalize enum fields
	gInput.OdHPrismDirection = normalizeEnumStr(gInput.OdHPrismDirection)
	gInput.OsHPrismDirection = normalizeEnumStr(gInput.OsHPrismDirection)
	gInput.OdVPrismDirection = normalizeEnumStr(gInput.OdVPrismDirection)
	gInput.OsVPrismDirection = normalizeEnumStr(gInput.OsVPrismDirection)
	cInput.OdType = normalizeEnumStr(cInput.OdType)
	cInput.OsType = normalizeEnumStr(cInput.OsType)

	// parse expiration dates
	gExp, err := parseExpDate(gInput.Date, gInput.ExpirationDate)
	if err != nil {
		return nil, errors.New("invalid glasses_data.expiration_date: " + err.Error())
	}
	cExp, err := parseExpDate(cInput.Date, cInput.ExpirationDate)
	if err != nil {
		return nil, errors.New("invalid contacts_data.expiration_date: " + err.Error())
	}

	glassesRec := presModel.GlassesPrescription{
		OdSph: gInput.OdSph, OsSph: gInput.OsSph,
		OdCyl: gInput.OdCyl, OsCyl: gInput.OsCyl,
		OdAxis: gInput.OdAxis, OsAxis: gInput.OsAxis,
		OdAdd: ffVal(gInput.OdAdd), OsAdd: ffVal(gInput.OsAdd),
		OdHPrism: ffVal(gInput.OdHPrism), OsHPrism: ffVal(gInput.OsHPrism),
		OdHPrismDirection: gInput.OdHPrismDirection, OsHPrismDirection: gInput.OsHPrismDirection,
		OdVPrism: ffVal(gInput.OdVPrism), OsVPrism: ffVal(gInput.OsVPrism),
		OdVPrismDirection: gInput.OdVPrismDirection, OsVPrismDirection: gInput.OsVPrismDirection,
		OdDpd: ffVal(gInput.OdDpd), OsDpd: ffVal(gInput.OsDpd),
		ExpirationDate: gExp,
	}

	contactsRec := presModel.ContactLensPrescription{
		OdContLens: cInput.OdContLens, OsContLens: cInput.OsContLens,
		OdBc: cInput.OdBc, OsBc: cInput.OsBc,
		OdDia: ffVal(cInput.OdDia), OsDia: ffVal(cInput.OsDia),
		OdPwr: cInput.OdPwr, OsPwr: cInput.OsPwr,
		OdCyl: cInput.OdCyl, OsCyl: cInput.OsCyl,
		OdAxis: cInput.OdAxis, OsAxis: cInput.OsAxis,
		OdAdd: cInput.OdAdd, OsAdd: cInput.OsAdd,
		OdColor: cInput.OdColor, OsColor: cInput.OsColor,
		OdType: cInput.OdType, OsType: cInput.OsType,
		ExpirationDate: cExp,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&pres).Error; err != nil {
			return err
		}

		glassesRec.PrescriptionID = pres.IDPatientPrescription
		contactsRec.PrescriptionID = pres.IDPatientPrescription

		if err := tx.Create(&glassesRec).Error; err != nil {
			return err
		}
		if err := tx.Create(&contactsRec).Error; err != nil {
			return err
		}

		// compute g_or_c
		gOrC := computeGOrC(&glassesRec, &contactsRec)
		return tx.Model(&pres).Update("g_or_c", gOrC).Error
	}); err != nil {
		return nil, err
	}

	date := ""
	if pres.PrescriptionDate != nil {
		date = pres.PrescriptionDate.Format("2006-01-02")
	}

	return map[string]interface{}{
		"id_rx":         pres.IDPatientPrescription,
		"date":          date,
		"g_or_c":        derefStr(pres.GOrC),
		"location":      locationShortName(loc),
		"patient_id":    pres.PatientID,
		"doctor":        pres.Doctor,
		"doctor_phone":  pres.PhoneNumber,
		"npi":           pres.NPI,
		"license":       pres.License,
		"type_location": locationFullName(loc),
		"prescription": map[string]interface{}{
			"glasses_data":  glassesToMap(&glassesRec),
			"contacts_data": contactsToMap(&contactsRec),
		},
	}, nil
}

// ─── PUT /rx ──────────────────────────────────────────────────────────────────

func (s *Service) UpdateRx(username string, input UpdateRxInput) error {
	if input.IDRx == 0 {
		return errors.New("id_rx is required")
	}

	var pres presModel.PatientPrescription
	if err := s.db.First(&pres, input.IDRx).Error; err != nil {
		return errors.New("prescription not found")
	}

	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil || loc == nil {
		return errors.New("employee or location not found")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// update main prescription fields
		presUpdates := map[string]interface{}{}

		if input.Date != nil && *input.Date != "" {
			t, err := time.Parse("2006-01-02", *input.Date)
			if err != nil {
				return errors.New("invalid date format")
			}
			presUpdates["prescription_date"] = t
		}
		if input.Note != nil         { presUpdates["note"] = *input.Note }
		if input.Medication != nil   { presUpdates["medication"] = *input.Medication }
		if input.Dosage != nil       { presUpdates["dosage"] = *input.Dosage }
		if input.DocumentLink != nil { presUpdates["document_link"] = *input.DocumentLink }
		if input.Doctor != nil       { presUpdates["doctor"] = *input.Doctor }
		if input.DoctorPhone != nil  { presUpdates["phone_number"] = *input.DoctorPhone }
		if input.NPI != nil          { presUpdates["npi"] = input.NPI }
		if input.License != nil      { presUpdates["license"] = input.License }
		if input.LocationID != nil   { presUpdates["location_id"] = loc.IDLocation }

		if input.Signature != nil && *input.Signature != "" {
			sig := *input.Signature
			if idx := strings.Index(sig, "/mnt/tank/data/"); idx >= 0 {
				sig = strings.TrimSpace(sig[idx+len("/mnt/tank/data/"):])
			}
			presUpdates["signature"] = sig
		}

		if len(presUpdates) > 0 {
			if err := tx.Model(&pres).Updates(presUpdates).Error; err != nil {
				return err
			}
		}

		// update glasses
		if input.Prescription != nil && input.Prescription.GlassesData != nil {
			gd := input.Prescription.GlassesData
			var g presModel.GlassesPrescription
			if tx.Where("prescription_id = ?", input.IDRx).First(&g).Error == nil {
				gExp, _ := parseExpDate(gd.Date, gd.ExpirationDate)

				gUpdates := map[string]interface{}{
					"od_sph":              gd.OdSph,
					"os_sph":              gd.OsSph,
					"od_cyl":              gd.OdCyl,
					"os_cyl":              gd.OsCyl,
					"od_axis":             gd.OdAxis,
					"os_axis":             gd.OsAxis,
					"od_add":              ffVal(gd.OdAdd),
					"os_add":              ffVal(gd.OsAdd),
					"od_h_prism":          ffVal(gd.OdHPrism),
					"os_h_prism":          ffVal(gd.OsHPrism),
					"od_h_prism_direction": normalizeEnumStr(gd.OdHPrismDirection),
					"os_h_prism_direction": normalizeEnumStr(gd.OsHPrismDirection),
					"od_v_prism":          ffVal(gd.OdVPrism),
					"os_v_prism":          ffVal(gd.OsVPrism),
					"od_v_prism_direction": normalizeEnumStr(gd.OdVPrismDirection),
					"os_v_prism_direction": normalizeEnumStr(gd.OsVPrismDirection),
					"od_dpd":              ffVal(gd.OdDpd),
					"os_dpd":              ffVal(gd.OsDpd),
					"expiration_date":      gExp,
				}
				if err := tx.Model(&g).Updates(gUpdates).Error; err != nil {
					return err
				}
			}
		}

		// update contacts
		if input.Prescription != nil && input.Prescription.ContactsData != nil {
			cd := input.Prescription.ContactsData
			var c presModel.ContactLensPrescription
			if tx.Where("prescription_id = ?", input.IDRx).First(&c).Error == nil {
				cExp, _ := parseExpDate(cd.Date, cd.ExpirationDate)

				cUpdates := map[string]interface{}{
					"od_cont_lens":   cd.OdContLens,
					"os_cont_lens":   cd.OsContLens,
					"od_bc":          cd.OdBc,
					"os_bc":          cd.OsBc,
					"od_dia":         ffVal(cd.OdDia),
					"os_dia":         ffVal(cd.OsDia),
					"od_pwr":         cd.OdPwr,
					"os_pwr":         cd.OsPwr,
					"od_cyl":         cd.OdCyl,
					"os_cyl":         cd.OsCyl,
					"od_axis":        cd.OdAxis,
					"os_axis":        cd.OsAxis,
					"od_add":         cd.OdAdd,
					"os_add":         cd.OsAdd,
					"od_color":       cd.OdColor,
					"os_color":       cd.OsColor,
					"od_type":        normalizeEnumStr(cd.OdType),
					"os_type":        normalizeEnumStr(cd.OsType),
					"expiration_date": cExp,
				}
				if err := tx.Model(&c).Updates(cUpdates).Error; err != nil {
					return err
				}
			}
		}

		// recompute g_or_c
		var g presModel.GlassesPrescription
		var c presModel.ContactLensPrescription
		tx.Where("prescription_id = ?", input.IDRx).First(&g)
		tx.Where("prescription_id = ?", input.IDRx).First(&c)

		hasG := hasGlassesData(&g)
		hasC := hasContactsData(&c)

		var newGOrC *string
		if hasG && hasC {
			v := "G, C"; newGOrC = &v
		} else if hasG {
			v := "G"; newGOrC = &v
		} else if hasC {
			v := "C"; newGOrC = &v
		}

		return tx.Model(&pres).Update("g_or_c", newGOrC).Error
	})
}

// ─── DELETE /rx/{id_rx} ───────────────────────────────────────────────────────

func (s *Service) DeleteRx(rxID int64) error {
	var pres presModel.PatientPrescription
	if err := s.db.First(&pres, rxID).Error; err != nil {
		return errors.New("prescription not found")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("prescription_id = ?", rxID).Delete(&presModel.GlassesPrescription{})
		tx.Where("prescription_id = ?", rxID).Delete(&presModel.ContactLensPrescription{})
		return tx.Delete(&pres).Error
	})
}


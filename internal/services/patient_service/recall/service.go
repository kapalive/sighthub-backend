package recall

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	locModel "sighthub-backend/internal/models/location"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

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
		return &emp, nil, errors.New("employee has no location")
	}
	var loc locModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return &emp, nil, err
	}
	return &emp, &loc, nil
}

// ─── Recall List ─────────────────────────────────────────────────────────────

type RecallListParams struct {
	// Pagination
	Page    int
	PerPage int

	// Recall filters
	DateFrom   *string // YYYY-MM-DD
	DateTo     *string // YYYY-MM-DD
	CallStatus *string // pending, reached, no_answer, unreachable, rescheduled
	Reason     *string

	// Patient filters (same as patient search)
	FirstName *string
	LastName  *string
	DOB       *string
	City      *string
	State     *string
	Phone     *string
	Email     *string

	// Special filters
	HasPhone            *bool  // true = only with phone, false = only without
	PreferredLanguageID *int   // filter by preferred_language_id
	InsuranceCompanyID  *int   // filter by insurance_company_id (via policy → holder)
	SortBy              string // date, patient_name (default: date)
	SortDir             string // asc, desc (default: asc)
}

type RecallListItem struct {
	RecallID      int64      `json:"recall_id"`
	Date          string     `json:"date"`
	Reason        *string    `json:"reason"`
	Note          *string    `json:"note"`
	CallStatus    *string    `json:"call_status"`
	CallAttempts  int        `json:"call_attempts"`
	LastAttemptAt *time.Time `json:"last_attempt_at"`
	SourceTable   *string    `json:"source_table"`

	// Patient info
	PatientID           int64   `json:"patient_id"`
	FirstName           string  `json:"first_name"`
	LastName            string  `json:"last_name"`
	Phone               *string `json:"phone"`
	PhoneHome           *string `json:"phone_home"`
	CellWork            *string `json:"cell_work"`
	Email               *string `json:"email"`
	DOB                 *string `json:"dob"`
	PreferredLanguageID *int64  `json:"preferred_language_id"`
	PreferredLanguage   *string `json:"preferred_language"`
	InsuranceCompany    *string `json:"insurance_company"`
}

type RecallListResult struct {
	Recalls    []RecallListItem `json:"recalls"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PerPage    int              `json:"per_page"`
	TotalPages int              `json:"total_pages"`
}

func (s *Service) GetRecallList(username string, params RecallListParams) (*RecallListResult, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, errors.New("employee or location not found")
	}

	// Defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 25
	}

	// Base query: join planing_communication with patient, language, insurance
	base := s.db.Table("planing_communication pc").
		Joins("JOIN patient p ON p.id_patient = pc.patient_id").
		Joins("LEFT JOIN preferred_language pl ON pl.id_preferred_language = p.preferred_language_id").
		Joins(`LEFT JOIN LATERAL (
			SELECT DISTINCT ON (ihp.patient_id) ic.id_insurance_company, ic.company_name
			FROM insurance_holder_patients ihp
			JOIN insurance_policy ip ON ip.id_insurance_policy = ihp.insurance_policy_id AND ip.active = true
			JOIN insurance_company ic ON ic.id_insurance_company = ip.insurance_company_id
			WHERE ihp.patient_id = p.id_patient
			ORDER BY ihp.patient_id, ihp.id_insurance_holder_patients ASC
			LIMIT 1
		) ins ON true`).
		Where("pc.location_id = ?", loc.IDLocation).
		Where("pc.date::date >= CURRENT_DATE")

	// ── Recall filters ──
	if params.DateFrom != nil {
		base = base.Where("pc.date::date >= ?", *params.DateFrom)
	}
	if params.DateTo != nil {
		base = base.Where("pc.date::date <= ?", *params.DateTo)
	}
	if params.CallStatus != nil {
		base = base.Where("COALESCE(pc.call_status, 'pending') = ?", *params.CallStatus)
	}
	if params.Reason != nil {
		base = base.Where("pc.reason ILIKE ?", "%"+*params.Reason+"%")
	}

	// ── Patient filters ──
	if params.FirstName != nil {
		base = base.Where("p.first_name ILIKE ?", *params.FirstName+"%")
	}
	if params.LastName != nil {
		base = base.Where("p.last_name ILIKE ?", *params.LastName+"%")
	}
	if params.DOB != nil {
		if t, err := time.Parse("2006-01-02", *params.DOB); err == nil {
			base = base.Where("p.dob = ?", t)
		}
	}
	if params.City != nil {
		base = base.Where("p.city ILIKE ?", *params.City+"%")
	}
	if params.State != nil {
		base = base.Where("p.state ILIKE ?", *params.State+"%")
	}
	if params.Phone != nil {
		base = base.Where("(p.phone ILIKE ? OR p.phone_home ILIKE ? OR p.cell_work ILIKE ?)",
			*params.Phone+"%", *params.Phone+"%", *params.Phone+"%")
	}
	if params.Email != nil {
		base = base.Where("p.email ILIKE ?", *params.Email+"%")
	}

	// ── Has phone filter ──
	if params.HasPhone != nil {
		if *params.HasPhone {
			base = base.Where("(p.phone IS NOT NULL AND p.phone != '' OR p.phone_home IS NOT NULL AND p.phone_home != '' OR p.cell_work IS NOT NULL AND p.cell_work != '')")
		} else {
			base = base.Where("(p.phone IS NULL OR p.phone = '') AND (p.phone_home IS NULL OR p.phone_home = '') AND (p.cell_work IS NULL OR p.cell_work = '')")
		}
	}

	// ── Preferred language filter ──
	if params.PreferredLanguageID != nil {
		base = base.Where("p.preferred_language_id = ?", *params.PreferredLanguageID)
	}

	// ── Insurance company filter ──
	if params.InsuranceCompanyID != nil {
		base = base.Where("ins.id_insurance_company = ?", *params.InsuranceCompanyID)
	}

	// ── Count total ──
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}

	// ── Sort ──
	sortBy := "pc.date"
	if strings.EqualFold(params.SortBy, "patient_name") {
		sortBy = "p.last_name"
	}
	sortDir := "ASC"
	if strings.EqualFold(params.SortDir, "desc") {
		sortDir = "DESC"
	}
	orderClause := fmt.Sprintf("%s %s", sortBy, sortDir)
	if sortBy == "p.last_name" {
		orderClause += fmt.Sprintf(", p.first_name %s", sortDir)
	}

	// ── Select + paginate ──
	type row struct {
		RecallID            int64      `gorm:"column:recall_id"`
		Date                time.Time  `gorm:"column:date"`
		Reason              *string    `gorm:"column:reason"`
		Note                *string    `gorm:"column:note"`
		CallStatus          *string    `gorm:"column:call_status"`
		CallAttempts        int        `gorm:"column:call_attempts"`
		LastAttemptAt       *time.Time `gorm:"column:last_attempt_at"`
		SourceTable         *string    `gorm:"column:source_table"`
		PatientID           int64      `gorm:"column:patient_id"`
		FirstName           string     `gorm:"column:first_name"`
		LastName            string     `gorm:"column:last_name"`
		Phone               *string    `gorm:"column:phone"`
		PhoneHome           *string    `gorm:"column:phone_home"`
		CellWork            *string    `gorm:"column:cell_work"`
		Email               *string    `gorm:"column:email"`
		DOB                 *time.Time `gorm:"column:dob"`
		PreferredLanguageID *int64     `gorm:"column:preferred_language_id"`
		PreferredLanguage   *string    `gorm:"column:language"`
		InsuranceCompany    *string    `gorm:"column:company_name"`
	}

	var rows []row
	offset := (params.Page - 1) * params.PerPage

	err = base.Select(`
		pc.id_planing_communication AS recall_id,
		pc.date,
		pc.reason,
		pc.note,
		pc.call_status,
		pc.call_attempts,
		pc.last_attempt_at,
		pc.source_table,
		p.id_patient AS patient_id,
		p.first_name,
		p.last_name,
		p.phone,
		p.phone_home,
		p.cell_work,
		p.email,
		p.dob,
		p.preferred_language_id,
		pl.language,
		ins.company_name
	`).
		Order(orderClause).
		Offset(offset).
		Limit(params.PerPage).
		Find(&rows).Error

	if err != nil {
		return nil, err
	}

	items := make([]RecallListItem, 0, len(rows))
	for _, r := range rows {
		item := RecallListItem{
			RecallID:            r.RecallID,
			Date:                r.Date.Format(time.RFC3339),
			Reason:              r.Reason,
			Note:                r.Note,
			CallStatus:          r.CallStatus,
			CallAttempts:        r.CallAttempts,
			LastAttemptAt:       r.LastAttemptAt,
			SourceTable:         r.SourceTable,
			PatientID:           r.PatientID,
			FirstName:           r.FirstName,
			LastName:            r.LastName,
			Phone:               r.Phone,
			PhoneHome:           r.PhoneHome,
			CellWork:            r.CellWork,
			Email:               r.Email,
			PreferredLanguageID: r.PreferredLanguageID,
			PreferredLanguage:   r.PreferredLanguage,
			InsuranceCompany:    r.InsuranceCompany,
		}
		if r.DOB != nil {
			dob := r.DOB.Format("2006-01-02")
			item.DOB = &dob
		}
		items = append(items, item)
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))

	return &RecallListResult{
		Recalls:    items,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}

// ─── Log Call Result ─────────────────────────────────────────────────────────

type LogCallResultInput struct {
	Status         string  `json:"status"` // reached, no_answer, unreachable, rescheduled
	Note           *string `json:"note"`
	RescheduleDate *string `json:"reschedule_date"` // YYYY-MM-DD or RFC3339
}

func (s *Service) LogCallResult(username string, recallID int64, input LogCallResultInput) error {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return errors.New("employee or location not found")
	}

	// Validate status
	validStatuses := map[string]bool{
		"reached":      true,
		"no_answer":    true,
		"unreachable":  true,
		"rescheduled":  true,
	}
	if !validStatuses[input.Status] {
		return errors.New("invalid status: must be reached, no_answer, unreachable, or rescheduled")
	}

	var recall struct {
		IDPlaningCommunication int64      `gorm:"column:id_planing_communication;primaryKey"`
		PatientID              int64      `gorm:"column:patient_id"`
		LocationID             *int64     `gorm:"column:location_id"`
		CallStatus             *string    `gorm:"column:call_status"`
		CallAttempts           int        `gorm:"column:call_attempts"`
		LastAttemptAt          *time.Time `gorm:"column:last_attempt_at"`
		LastAttemptBy          *int64     `gorm:"column:last_attempt_by"`
		Note                   *string    `gorm:"column:note"`
		Date                   time.Time  `gorm:"column:date"`
	}

	if err := s.db.Table("planing_communication").
		Where("id_planing_communication = ? AND location_id = ?", recallID, loc.IDLocation).
		First(&recall).Error; err != nil {
		return errors.New("recall not found")
	}

	now := time.Now()
	empID := emp.IDEmployee
	updates := map[string]interface{}{
		"call_status":     input.Status,
		"call_attempts":   recall.CallAttempts + 1,
		"last_attempt_at": now,
		"last_attempt_by": empID,
	}

	if input.Note != nil {
		updates["note"] = *input.Note
	}

	// If rescheduling, update the recall date
	if input.Status == "rescheduled" {
		if input.RescheduleDate == nil || *input.RescheduleDate == "" {
			return errors.New("reschedule_date is required when status is 'rescheduled'")
		}
		newDate, err := parseDate(*input.RescheduleDate)
		if err != nil {
			return errors.New("invalid reschedule_date format")
		}
		updates["date"] = newDate
		// Reset status to pending for the new date
		updates["call_status"] = "pending"
		updates["call_attempts"] = 0
	}

	return s.db.Table("planing_communication").
		Where("id_planing_communication = ?", recallID).
		Updates(updates).Error
}

func parseDate(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			if layout == "2006-01-02" {
				// Set to 14:00 UTC (09:00 NY)
				t = time.Date(t.Year(), t.Month(), t.Day(), 14, 0, 0, 0, time.UTC)
			}
			return t, nil
		}
	}
	return time.Time{}, errors.New("invalid date")
}

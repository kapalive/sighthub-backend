package questionnaire_service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	mktModel "sighthub-backend/internal/models/marketing"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type CreateReferralInput struct {
	PatientID         *int64 `json:"patient_id"`
	VisitReasonsID    *int   `json:"visit_reasons_id"`
	ReferralSourcesID *int   `json:"referral_sources_id"`
	// set by handler
	LocationID int
	EmployeeID int
}

type ReferralSourceItem struct {
	ReferralSourcesID int    `json:"referral_sources_id"`
	Title             string `json:"title"`
}

type VisitReasonItem struct {
	VisitReasonsID int    `json:"visit_reasons_id"`
	Title          string `json:"title"`
}

// ─── Employee/location helper ─────────────────────────────────────────────────

func (s *Service) GetEmployeeAndLocation(username string) (*empModel.Employee, int, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	locID := 0
	if emp.LocationID != nil {
		locID = int(*emp.LocationID)
	}
	return &emp, locID, nil
}

// ─── Endpoints ────────────────────────────────────────────────────────────────

func (s *Service) CreateReferral(input CreateReferralInput) (*mktModel.QuestionnaireReferral, error) {
	if input.PatientID == nil || input.VisitReasonsID == nil || input.ReferralSourcesID == nil {
		return nil, fmt.Errorf("missing required fields")
	}

	now := time.Now()
	empID := int64(input.EmployeeID)
	rec := &mktModel.QuestionnaireReferral{
		PatientID:         input.PatientID,
		VisitReasonsID:    input.VisitReasonsID,
		ReferralSourcesID: input.ReferralSourcesID,
		LocationID:        &input.LocationID,
		EmployeeID:        &empID,
		DatetimeCreated:   &now,
	}
	if err := s.db.Create(rec).Error; err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *Service) GetReferralSources() ([]ReferralSourceItem, error) {
	var rows []mktModel.ReferralSource
	if err := s.db.Order("title ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ReferralSourceItem, len(rows))
	for i, r := range rows {
		out[i] = ReferralSourceItem{ReferralSourcesID: r.IDReferralSources, Title: r.Title}
	}
	return out, nil
}

func (s *Service) GetVisitReasons() ([]VisitReasonItem, error) {
	var rows []mktModel.VisitReason
	if err := s.db.Order("title ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]VisitReasonItem, len(rows))
	for i, r := range rows {
		out[i] = VisitReasonItem{VisitReasonsID: r.IDVisitReasons, Title: r.Title}
	}
	return out, nil
}

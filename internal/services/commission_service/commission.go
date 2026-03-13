package commission_service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	empModel "sighthub-backend/internal/models/employees"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

type CommissionItem struct {
	IDEmployeeCommissions int     `json:"id_employee_commissions"`
	EmployeeID            int     `json:"employee_id"`
	StartDate             string  `json:"start_date"`
	EndDate               *string `json:"end_date"`
	CommissionPercent     float64 `json:"commission_percent"`
	CreatedAt             *string `json:"created_at"`
	UpdatedAt             *string `json:"updated_at"`
}

type CommissionDetail struct {
	PBKey        string   `json:"pb_key"`
	PercentValue float64  `json:"percent_value"`
	SumMin       *float64 `json:"sum_min"`
	BrandID      *int     `json:"brand_id"`
}

type CurrentCommissionResult struct {
	StartDate         *string            `json:"start_date"`
	EndDate           *string            `json:"end_date"`
	CommissionPercent *float64           `json:"commission_percent"`
	Details           map[string]float64 `json:"details"`
}

type HistoryItem struct {
	IDEmployeeHistoryCommissions int                `json:"id_employee_history_commissions"`
	EmployeeCommissionsID        int                `json:"employee_commissions_id"`
	EmployeeID                   int                `json:"employee_id"`
	StartDate                    string             `json:"start_date"`
	EndDate                      *string            `json:"end_date"`
	CommissionPercent            float64            `json:"commission_percent"`
	CreatedAt                    *string            `json:"created_at"`
	Details                      map[string]float64 `json:"details"`
}

type CreateCommissionInput struct {
	StartDate         string             `json:"start_date"`
	EndDate           string             `json:"end_date"`
	CommissionPercent float64            `json:"commission_percent"`
	Details           map[string]float64 `json:"details"`
}

type UpdateCommissionInput struct {
	StartDate         string             `json:"start_date"`
	EndDate           string             `json:"end_date"`
	CommissionPercent float64            `json:"commission_percent"`
	Details           []CommissionDetail `json:"details"`
}

func (s *Service) GetCommissions(employeeID int) ([]CommissionItem, error) {
	var list []empModel.EmployeeCommissions
	if err := s.db.Where("employee_id = ?", employeeID).Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]CommissionItem, 0, len(list))
	for _, c := range list {
		result = append(result, toCommissionItem(c))
	}
	return result, nil
}

func (s *Service) GetCurrentCommission(employeeID int) (*CurrentCommissionResult, error) {
	var c empModel.EmployeeCommissions
	err := s.db.Where("employee_id = ?", employeeID).Order("created_at DESC").First(&c).Error
	if err != nil {
		return &CurrentCommissionResult{
			Details: map[string]float64{
				"Frames": 0, "Lens": 0, "Contact Lens": 0, "Prof. service": 0, "Add service": 0,
			},
		}, nil
	}
	var rels []empModel.EmployeeCommissionsDetailsRelation
	s.db.Where("employee_commissions_id = ?", c.IDEmployeeCommissions).Find(&rels)
	details := map[string]float64{}
	for _, rel := range rels {
		var det empModel.EmployeeCommissionsDetails
		if s.db.Where("id_details = ?", rel.DetailsID).First(&det).Error == nil && det.PBKey != "" {
			details[det.PBKey] = det.PercentValue
		}
	}
	start := c.StartDate.Format("2006-01-02")
	result := &CurrentCommissionResult{StartDate: &start, CommissionPercent: &c.CommissionPercent, Details: details}
	if !c.EndDate.IsZero() {
		end := c.EndDate.Format("2006-01-02")
		result.EndDate = &end
	}
	return result, nil
}

func (s *Service) GetCommissionHistory(employeeID int) ([]HistoryItem, error) {
	var history []empModel.EmployeeHistoryCommissions
	if err := s.db.Where("employee_id = ?", employeeID).Order("created_at DESC").Find(&history).Error; err != nil {
		return nil, err
	}
	result := make([]HistoryItem, 0, len(history))
	for _, h := range history {
		var rels []empModel.EmployeeHistoryCommissionsDetailsRelation
		s.db.Where("history_commissions_id = ?", h.IDEmployeeHistoryCommissions).Find(&rels)
		details := map[string]float64{}
		for _, rel := range rels {
			var det empModel.EmployeeCommissionsDetails
			if s.db.Where("id_details = ?", rel.DetailsID).First(&det).Error == nil && det.PBKey != "" {
				details[det.PBKey] = det.PercentValue
			}
		}
		item := HistoryItem{
			IDEmployeeHistoryCommissions: h.IDEmployeeHistoryCommissions,
			EmployeeCommissionsID:        h.EmployeeCommissionsID,
			EmployeeID:                   h.EmployeeID,
			StartDate:                    h.StartDate.Format("2006-01-02"),
			CommissionPercent:            h.CommissionPercent,
			Details:                      details,
		}
		if h.EndDate != nil && !h.EndDate.IsZero() {
			end := h.EndDate.Format("2006-01-02")
			item.EndDate = &end
		}
		if h.CreatedAt != nil {
			ts := h.CreatedAt.Format(time.RFC3339)
			item.CreatedAt = &ts
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) UpdateCommission(employeeID, commissionID int, input UpdateCommissionInput) error {
	var commission empModel.EmployeeCommissions
	if err := s.db.Where("id_employee_commissions = ? AND employee_id = ?", commissionID, employeeID).First(&commission).Error; err != nil {
		return errors.New("commission not found")
	}
	if input.StartDate == "" {
		return errors.New("start_date and commission_percent are required")
	}
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return errors.New("invalid date format, use YYYY-MM-DD")
	}
	var endDate time.Time
	if input.EndDate != "" {
		if endDate, err = time.Parse("2006-01-02", input.EndDate); err != nil {
			return errors.New("invalid end_date format")
		}
		if startDate.After(endDate) {
			return fmt.Errorf("start date cannot be after end date")
		}
	}
	if input.CommissionPercent < 0 || input.CommissionPercent > 100 {
		return errors.New("commission percent must be between 0 and 100")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		hist := empModel.EmployeeHistoryCommissions{
			EmployeeCommissionsID: commission.IDEmployeeCommissions,
			EmployeeID:            commission.EmployeeID,
			StartDate:             commission.StartDate,
			CommissionPercent:     commission.CommissionPercent,
			CreatedAt:             commission.UpdatedAt,
		}
		if !commission.EndDate.IsZero() {
			hist.EndDate = &commission.EndDate
		}
		if err := tx.Create(&hist).Error; err != nil {
			return err
		}
		var rels []empModel.EmployeeCommissionsDetailsRelation
		tx.Where("employee_commissions_id = ?", commission.IDEmployeeCommissions).Find(&rels)
		for _, rel := range rels {
			tx.Create(&empModel.EmployeeHistoryCommissionsDetailsRelation{HistoryCommissionsID: hist.IDEmployeeHistoryCommissions, DetailsID: rel.DetailsID})
		}
		commission.StartDate = startDate
		commission.EndDate = endDate
		commission.CommissionPercent = input.CommissionPercent
		if err := tx.Save(&commission).Error; err != nil {
			return err
		}
		tx.Where("employee_commissions_id = ?", commission.IDEmployeeCommissions).Delete(&empModel.EmployeeCommissionsDetailsRelation{})
		for _, d := range input.Details {
			det := empModel.EmployeeCommissionsDetails{PBKey: d.PBKey, PercentValue: d.PercentValue, SumMin: d.SumMin, BrandID: d.BrandID}
			if err := tx.Create(&det).Error; err != nil {
				return err
			}
			tx.Create(&empModel.EmployeeCommissionsDetailsRelation{EmployeeCommissionsID: commission.IDEmployeeCommissions, DetailsID: det.IDDetails})
		}
		return nil
	})
}

func (s *Service) CreateCommission(employeeID int, input CreateCommissionInput) (int, error) {
	if input.StartDate == "" {
		return 0, errors.New("start_date and commission_percent are required")
	}
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return 0, errors.New("invalid date format, use YYYY-MM-DD")
	}
	var endDate time.Time
	if input.EndDate != "" {
		if endDate, err = time.Parse("2006-01-02", input.EndDate); err != nil {
			return 0, errors.New("invalid end_date format")
		}
		if startDate.After(endDate) {
			return 0, errors.New("start date cannot be after end date")
		}
	}
	if input.CommissionPercent < 0 || input.CommissionPercent > 100 {
		return 0, errors.New("commission percent must be between 0 and 100")
	}
	var id int
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var latest empModel.EmployeeCommissions
		if tx.Where("employee_id = ?", employeeID).Order("created_at DESC").First(&latest).Error == nil {
			today := time.Now().UTC().Truncate(24 * time.Hour)
			latest.EndDate = today
			hist := empModel.EmployeeHistoryCommissions{
				EmployeeCommissionsID: latest.IDEmployeeCommissions,
				EmployeeID:            latest.EmployeeID,
				StartDate:             latest.StartDate,
				EndDate:               &today,
				CommissionPercent:     latest.CommissionPercent,
				CreatedAt:             latest.UpdatedAt,
			}
			if err := tx.Create(&hist).Error; err != nil {
				return err
			}
			var rels []empModel.EmployeeCommissionsDetailsRelation
			tx.Where("employee_commissions_id = ?", latest.IDEmployeeCommissions).Find(&rels)
			for _, rel := range rels {
				tx.Create(&empModel.EmployeeHistoryCommissionsDetailsRelation{HistoryCommissionsID: hist.IDEmployeeHistoryCommissions, DetailsID: rel.DetailsID})
			}
			tx.Save(&latest)
		}
		commission := empModel.EmployeeCommissions{EmployeeID: employeeID, StartDate: startDate, EndDate: endDate, CommissionPercent: input.CommissionPercent}
		if err := tx.Create(&commission).Error; err != nil {
			return err
		}
		id = commission.IDEmployeeCommissions
		for pbKey, pctVal := range input.Details {
			det := empModel.EmployeeCommissionsDetails{PBKey: pbKey, PercentValue: pctVal}
			if err := tx.Create(&det).Error; err != nil {
				return err
			}
			tx.Create(&empModel.EmployeeCommissionsDetailsRelation{EmployeeCommissionsID: commission.IDEmployeeCommissions, DetailsID: det.IDDetails})
		}
		return nil
	})
	return id, err
}

func toCommissionItem(c empModel.EmployeeCommissions) CommissionItem {
	item := CommissionItem{IDEmployeeCommissions: c.IDEmployeeCommissions, EmployeeID: c.EmployeeID, StartDate: c.StartDate.Format("2006-01-02"), CommissionPercent: c.CommissionPercent}
	if !c.EndDate.IsZero() {
		end := c.EndDate.Format("2006-01-02")
		item.EndDate = &end
	}
	if c.CreatedAt != nil {
		ts := c.CreatedAt.Format(time.RFC3339); item.CreatedAt = &ts
	}
	if c.UpdatedAt != nil {
		ts := c.UpdatedAt.Format(time.RFC3339); item.UpdatedAt = &ts
	}
	return item
}

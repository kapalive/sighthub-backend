package employees_repo

import (
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

// ─────────────────────────────────────────────
// DTO types
// ─────────────────────────────────────────────

type CommissionDetail struct {
	PBKey        string   `json:"pb_key"`
	PercentValue float64  `json:"percent_value"`
	SumMin       *float64 `json:"sum_min"`
	BrandID      *int     `json:"brand_id"`
}

type CommissionWithDetails struct {
	IDEmployeeCommissions int                `json:"id_employee_commissions"`
	EmployeeID            int                `json:"employee_id"`
	StartDate             string             `json:"start_date"`
	EndDate               string             `json:"end_date"`
	CommissionPercent     float64            `json:"commission_percent"`
	Details               map[string]float64 `json:"details"`
}

type CreateCommissionInput struct {
	StartDate         string                      `json:"start_date"`
	EndDate           string                      `json:"end_date"`
	CommissionPercent float64                     `json:"commission_percent"`
	Details           map[string]CommissionDetail `json:"details"`
}

// ─────────────────────────────────────────────
// CommissionRepo
// ─────────────────────────────────────────────

type CommissionRepo struct {
	DB *gorm.DB
}

func NewCommissionRepo(db *gorm.DB) *CommissionRepo {
	return &CommissionRepo{DB: db}
}

// GetAll returns all commission records for the given employee.
func (r *CommissionRepo) GetAll(employeeID int) ([]employees.EmployeeCommissions, error) {
	var commissions []employees.EmployeeCommissions
	if err := r.DB.
		Where("employee_id = ?", employeeID).
		Order("start_date DESC").
		Find(&commissions).Error; err != nil {
		return nil, err
	}
	return commissions, nil
}

// GetCurrent returns the most recent commission for the employee with its details.
func (r *CommissionRepo) GetCurrent(employeeID int) (*CommissionWithDetails, error) {
	var commission employees.EmployeeCommissions
	if err := r.DB.
		Where("employee_id = ?", employeeID).
		Order("start_date DESC").
		First(&commission).Error; err != nil {
		return nil, err
	}
	return r.loadWithDetails(&commission)
}

// GetHistory returns archived commissions with their details.
func (r *CommissionRepo) GetHistory(employeeID int) ([]CommissionWithDetails, error) {
	var historyRows []employees.EmployeeHistoryCommissions
	if err := r.DB.
		Where("employee_id = ?", employeeID).
		Order("created_at DESC").
		Find(&historyRows).Error; err != nil {
		return nil, err
	}

	// Collect unique commission ids from history to return their details
	seen := make(map[int]struct{}, len(historyRows))
	commissionIDs := make([]int, 0, len(historyRows))
	for _, h := range historyRows {
		if _, ok := seen[h.EmployeeCommissionsID]; !ok {
			seen[h.EmployeeCommissionsID] = struct{}{}
			commissionIDs = append(commissionIDs, h.EmployeeCommissionsID)
		}
	}

	var commissions []employees.EmployeeCommissions
	if len(commissionIDs) > 0 {
		if err := r.DB.Where("id_employee_commissions IN ?", commissionIDs).Find(&commissions).Error; err != nil {
			return nil, err
		}
	}

	result := make([]CommissionWithDetails, 0, len(commissions))
	for _, c := range commissions {
		cwd, err := r.loadWithDetails(&c)
		if err != nil {
			return nil, err
		}
		result = append(result, *cwd)
	}
	return result, nil
}

// Create archives the current commission (if any), closes it, and creates a new one
// with the supplied details — all within a single transaction.
// Returns the id of the newly created EmployeeCommissions record.
func (r *CommissionRepo) Create(employeeID int, input CreateCommissionInput) (int, error) {
	var newID int

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// Find the latest active commission
		var current employees.EmployeeCommissions
		hasExisting := false
		if err := tx.Where("employee_id = ?", employeeID).
			Order("start_date DESC").
			First(&current).Error; err == nil {
			hasExisting = true
		}

		if hasExisting {
			// Archive to history
			if err := archiveCommission(tx, &current); err != nil {
				return err
			}
			// Close old commission
			today := time.Now().Truncate(24 * time.Hour)
			if err := tx.Model(&current).Update("end_date", today).Error; err != nil {
				return err
			}
		}

		// Parse dates
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return err
		}
		endDate, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return err
		}

		// Create new EmployeeCommissions
		now := time.Now()
		newCommission := employees.EmployeeCommissions{
			EmployeeID:        employeeID,
			StartDate:         startDate,
			EndDate:           endDate,
			CommissionPercent: input.CommissionPercent,
			CreatedAt:         &now,
			UpdatedAt:         &now,
		}
		if err := tx.Create(&newCommission).Error; err != nil {
			return err
		}
		newID = newCommission.IDEmployeeCommissions

		// Create details + relations
		return createCommissionDetails(tx, newCommission.IDEmployeeCommissions, input.Details)
	})
	if err != nil {
		return 0, err
	}
	return newID, nil
}

// Update archives the current state of the commission, then replaces its data
// with the values from input — all within a single transaction.
func (r *CommissionRepo) Update(employeeID, commissionID int, input CreateCommissionInput) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var commission employees.EmployeeCommissions
		if err := tx.First(&commission, "id_employee_commissions = ? AND employee_id = ?",
			commissionID, employeeID).Error; err != nil {
			return err
		}

		// Archive current state
		if err := archiveCommission(tx, &commission); err != nil {
			return err
		}

		// Parse dates
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return err
		}
		endDate, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return err
		}

		now := time.Now()
		if err := tx.Model(&commission).Updates(map[string]interface{}{
			"start_date":         startDate,
			"end_date":           endDate,
			"commission_percent": input.CommissionPercent,
			"updated_at":         now,
		}).Error; err != nil {
			return err
		}

		// Delete old detail relations and their detail rows
		var oldRelations []employees.EmployeeCommissionsDetailsRelation
		tx.Where("employee_commissions_id = ?", commissionID).Find(&oldRelations)
		for _, rel := range oldRelations {
			tx.Delete(&employees.EmployeeCommissionsDetails{}, "id_details = ?", rel.DetailsID)
		}
		tx.Where("employee_commissions_id = ?", commissionID).
			Delete(&employees.EmployeeCommissionsDetailsRelation{})

		// Create new details
		return createCommissionDetails(tx, commissionID, input.Details)
	})
}

// ─────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────

// loadWithDetails builds a CommissionWithDetails from an EmployeeCommissions record
// by loading the related detail rows via the relation table.
func (r *CommissionRepo) loadWithDetails(c *employees.EmployeeCommissions) (*CommissionWithDetails, error) {
	var relations []employees.EmployeeCommissionsDetailsRelation
	if err := r.DB.Where("employee_commissions_id = ?", c.IDEmployeeCommissions).
		Find(&relations).Error; err != nil {
		return nil, err
	}

	detailsMap := make(map[string]float64, len(relations))
	for _, rel := range relations {
		var detail employees.EmployeeCommissionsDetails
		if err := r.DB.First(&detail, "id_details = ?", rel.DetailsID).Error; err == nil {
			detailsMap[detail.PBKey] = detail.PercentValue
		}
	}

	return &CommissionWithDetails{
		IDEmployeeCommissions: c.IDEmployeeCommissions,
		EmployeeID:            c.EmployeeID,
		StartDate:             c.StartDate.Format("2006-01-02"),
		EndDate:               c.EndDate.Format("2006-01-02"),
		CommissionPercent:     c.CommissionPercent,
		Details:               detailsMap,
	}, nil
}

// archiveCommission copies a commission and its details to the history tables.
func archiveCommission(tx *gorm.DB, c *employees.EmployeeCommissions) error {
	now := time.Now()
	hist := employees.EmployeeHistoryCommissions{
		EmployeeCommissionsID: c.IDEmployeeCommissions,
		EmployeeID:            c.EmployeeID,
		StartDate:             c.StartDate,
		EndDate:               &c.EndDate,
		CommissionPercent:     c.CommissionPercent,
		CreatedAt:             &now,
	}
	if err := tx.Create(&hist).Error; err != nil {
		return err
	}

	// Copy detail relations to history
	var relations []employees.EmployeeCommissionsDetailsRelation
	tx.Where("employee_commissions_id = ?", c.IDEmployeeCommissions).Find(&relations)
	for _, rel := range relations {
		histRel := employees.EmployeeHistoryCommissionsDetailsRelation{
			HistoryCommissionsID: hist.IDEmployeeHistoryCommissions,
			DetailsID:            rel.DetailsID,
		}
		if err := tx.Create(&histRel).Error; err != nil {
			return err
		}
	}
	return nil
}

// createCommissionDetails inserts EmployeeCommissionsDetails and their relations
// for the given commissionID.
func createCommissionDetails(tx *gorm.DB, commissionID int, details map[string]CommissionDetail) error {
	for _, d := range details {
		detail := employees.EmployeeCommissionsDetails{
			PBKey:        d.PBKey,
			PercentValue: d.PercentValue,
			SumMin:       d.SumMin,
			BrandID:      d.BrandID,
		}
		if err := tx.Create(&detail).Error; err != nil {
			return err
		}
		rel := employees.EmployeeCommissionsDetailsRelation{
			EmployeeCommissionsID: commissionID,
			DetailsID:             detail.IDDetails,
		}
		if err := tx.Create(&rel).Error; err != nil {
			return err
		}
	}
	return nil
}

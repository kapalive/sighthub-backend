package employees_repo

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/audit"
	"sighthub-backend/internal/models/employees"
)

// ─────────────────────────────────────────────
// DTO types
// ─────────────────────────────────────────────

type TimecardListItem struct {
	ID         int        `json:"id"`
	EmployeeID *int       `json:"employee_id"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	Active     bool       `json:"active"`
	Username   string     `json:"username"`
	LastAction string     `json:"last_action"`
	Timestamp  *time.Time `json:"timestamp"`
}

type TimecardPeriod struct {
	Date     string     `json:"date"`
	Checkin  *time.Time `json:"checkin"`
	Checkout *time.Time `json:"checkout"`
	Summary  string     `json:"summary"`
	Note     *string    `json:"note"`
}

type TimecardHistoryResult struct {
	TotalTime string           `json:"total_time"`
	Periods   []TimecardPeriod `json:"periods"`
}

type CreateTimecardInput struct {
	Username   string  `json:"username"`
	Password   string  `json:"password"`
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	EmployeeID *int    `json:"employee_id"`
}

type UpdateTimecardInput struct {
	Username  *string `json:"username"`
	Password  *string `json:"password"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}

// ─────────────────────────────────────────────
// TimecardRepo
// ─────────────────────────────────────────────

type TimecardRepo struct {
	DB *gorm.DB
}

func NewTimecardRepo(db *gorm.DB) *TimecardRepo {
	return &TimecardRepo{DB: db}
}

// GetList returns all timecard users with their last recorded action.
func (r *TimecardRepo) GetList() ([]TimecardListItem, error) {
	var logins []employees.EmployeeTimecardLogin
	if err := r.DB.Find(&logins).Error; err != nil {
		return nil, err
	}

	items := make([]TimecardListItem, 0, len(logins))
	for _, l := range logins {
		var last audit.EmployeeTimecardHistory
		err := r.DB.
			Where("employee_timecard_login_id = ?", l.IDEmployeeTimecardLogin).
			Order("timestamp DESC").
			First(&last).Error

		item := TimecardListItem{
			ID:         l.IDEmployeeTimecardLogin,
			EmployeeID: l.EmployeeID,
			FirstName:  l.FirstName,
			LastName:   l.LastName,
			Active:     l.Active,
			Username:   l.Username,
		}
		if err == nil {
			item.LastAction = last.ActionType
			item.Timestamp = &last.Timestamp
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetHistory returns timecard history for the given timecard login, filtered
// by the supplied date range. Events are paired into checkin/checkout periods.
func (r *TimecardRepo) GetHistory(timecardLoginID int, dateFrom, dateTo time.Time) (*TimecardHistoryResult, error) {
	var events []audit.EmployeeTimecardHistory
	err := r.DB.
		Where("employee_timecard_login_id = ?", timecardLoginID).
		Where("timestamp >= ? AND timestamp < ?", dateFrom, dateTo.Add(24*time.Hour)).
		Order("timestamp ASC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return buildTimecardHistory(events), nil
}

// GetHistoryForEmployee returns timecard history for the employee whose
// EmployeeTimecardLogin has the given employee_id FK.
func (r *TimecardRepo) GetHistoryForEmployee(employeeID int, dateFrom, dateTo time.Time) (*TimecardHistoryResult, error) {
	var login employees.EmployeeTimecardLogin
	if err := r.DB.Where("employee_id = ?", employeeID).First(&login).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &TimecardHistoryResult{TotalTime: "0h 0m", Periods: []TimecardPeriod{}}, nil
		}
		return nil, err
	}
	return r.GetHistory(login.IDEmployeeTimecardLogin, dateFrom, dateTo)
}

// Create creates a new timecard login record.
// When EmployeeID is provided and FirstName/LastName are nil, the names are
// looked up from the linked Employee record.
func (r *TimecardRepo) Create(input CreateTimecardInput) (int, error) {
	firstName := ""
	lastName := ""

	if input.FirstName != nil {
		firstName = *input.FirstName
	}
	if input.LastName != nil {
		lastName = *input.LastName
	}

	// Auto-fill names from employee when not provided
	if input.EmployeeID != nil && (firstName == "" || lastName == "") {
		var emp employees.Employee
		if err := r.DB.First(&emp, "id_employee = ?", *input.EmployeeID).Error; err == nil {
			if firstName == "" {
				firstName = emp.FirstName
			}
			if lastName == "" {
				lastName = emp.LastName
			}
		}
	}

	login := &employees.EmployeeTimecardLogin{
		Username:   input.Username,
		FirstName:  firstName,
		LastName:   lastName,
		EmployeeID: input.EmployeeID,
		Active:     true,
	}
	if err := login.SetPassword(input.Password); err != nil {
		return 0, err
	}
	if err := r.DB.Create(login).Error; err != nil {
		return 0, err
	}
	return login.IDEmployeeTimecardLogin, nil
}

// Update updates mutable fields on the timecard login identified by employeeID.
func (r *TimecardRepo) Update(employeeID int, input UpdateTimecardInput) error {
	var login employees.EmployeeTimecardLogin
	if err := r.DB.Where("employee_id = ?", employeeID).First(&login).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{}
	if input.Username != nil {
		updates["username"] = *input.Username
	}
	if input.FirstName != nil {
		updates["first_name"] = *input.FirstName
	}
	if input.LastName != nil {
		updates["last_name"] = *input.LastName
	}
	if input.Password != nil {
		if err := login.SetPassword(*input.Password); err != nil {
			return err
		}
		updates["password_hash"] = login.PasswordHash
	}

	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&login).Updates(updates).Error
}

// Deactivate sets active=false on the timecard login with the given primary key.
func (r *TimecardRepo) Deactivate(loginID int) error {
	return r.DB.Model(&employees.EmployeeTimecardLogin{}).
		Where("id_employee_timecard_login = ?", loginID).
		Update("active", false).Error
}

// GetByTimecardLoginID returns the timecard login by its primary key.
func (r *TimecardRepo) GetByTimecardLoginID(id int) (*employees.EmployeeTimecardLogin, error) {
	var login employees.EmployeeTimecardLogin
	if err := r.DB.First(&login, "id_employee_timecard_login = ?", id).Error; err != nil {
		return nil, err
	}
	return &login, nil
}

// ─────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────

// buildTimecardHistory pairs checkin/checkout events into periods and sums durations.
func buildTimecardHistory(events []audit.EmployeeTimecardHistory) *TimecardHistoryResult {
	periods := make([]TimecardPeriod, 0)
	totalMinutes := 0

	i := 0
	for i < len(events) {
		ev := events[i]
		if ev.ActionType != "checkin" {
			i++
			continue
		}

		checkinTime := ev.Timestamp
		dateStr := checkinTime.Format("2006-01-02")

		period := TimecardPeriod{
			Date:    dateStr,
			Checkin: &checkinTime,
			Note:    ev.Note,
		}

		// Look for the next checkout
		if i+1 < len(events) && events[i+1].ActionType == "checkout" {
			checkoutTime := events[i+1].Timestamp
			period.Checkout = &checkoutTime
			duration := int(checkoutTime.Sub(checkinTime).Minutes())
			if duration < 0 {
				duration = 0
			}
			totalMinutes += duration
			period.Summary = formatMinutes(duration)
			i += 2
		} else {
			period.Summary = "in progress"
			i++
		}

		periods = append(periods, period)
	}

	return &TimecardHistoryResult{
		TotalTime: formatMinutes(totalMinutes),
		Periods:   periods,
	}
}

func formatMinutes(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	return fmt.Sprintf("%dh %dm", h, m)
}

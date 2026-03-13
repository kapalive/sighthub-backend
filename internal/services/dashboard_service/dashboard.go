package dashboard_service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	locModel "sighthub-backend/internal/models/location"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("username = ?", username).First(&login).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	if emp.LocationID == nil {
		return nil, nil, errors.New("employee or location not found")
	}
	var loc locModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	return &emp, &loc, nil
}

// ── GetWeeklyIncome ─────────────────────────────────────────────────────────

func (s *Service) GetWeeklyIncome(username string) ([]map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	weekday := int(today.Weekday())
	// Python weekday(): Monday=0. Go Weekday(): Sunday=0.
	// Adjust to start week on Monday.
	offset := weekday - 1
	if offset < 0 {
		offset = 6
	}
	startOfWeek := today.AddDate(0, 0, -offset)
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	type dayRow struct {
		Day   time.Time
		Total float64
	}
	var rows []dayRow
	s.db.Raw(`
		SELECT DATE(ph.payment_timestamp) as day, SUM(ph.amount) as total
		FROM payment_history ph
		JOIN invoice i ON ph.invoice_id = i.id_invoice
		WHERE i.location_id = ?
		  AND DATE(ph.payment_timestamp) >= ?
		  AND DATE(ph.payment_timestamp) <= ?
		GROUP BY DATE(ph.payment_timestamp)
	`, loc.IDLocation, startOfWeek.Format("2006-01-02"), endOfWeek.Format("2006-01-02")).Scan(&rows)

	salesByDate := map[string]float64{}
	for _, r := range rows {
		salesByDate[r.Day.Format("2006-01-02")] = r.Total
	}

	dayNames := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	var data []map[string]interface{}
	for i := 0; i < 7; i++ {
		dayDate := startOfWeek.AddDate(0, 0, i)
		key := dayDate.Format("2006-01-02")
		total := salesByDate[key]
		data = append(data, map[string]interface{}{
			"day":   dayNames[i],
			"total": roundTo2(total),
		})
	}

	return data, nil
}

// ── GetAppointmentStatuses ──────────────────────────────────────────────────

func (s *Service) GetAppointmentStatuses(username string) ([]map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	type statusRow struct {
		Label string
		Total int
	}
	var rows []statusRow
	s.db.Raw(`
		SELECT sa.status_appointment as label, COUNT(a.id_appointment) as total
		FROM status_appointment sa
		LEFT JOIN appointment a ON a.status_appointment_id = sa.id_status_appointment
		  AND a.location_id = ?
		GROUP BY sa.id_status_appointment, sa.status_appointment
	`, loc.IDLocation).Scan(&rows)

	var data []map[string]interface{}
	for _, r := range rows {
		data = append(data, map[string]interface{}{
			"label": r.Label,
			"total": r.Total,
		})
	}

	return data, nil
}

// ── GetEmployeeSales ────────────────────────────────────────────────────────

func (s *Service) GetEmployeeSales(username, period string) ([]map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	startDate, endDate := periodDates(period)

	type salesRow struct {
		Name  string
		Sales float64
	}
	var rows []salesRow
	s.db.Raw(`
		SELECT CONCAT(e.first_name, ' ', e.last_name) as name, SUM(ph.amount) as sales
		FROM employee e
		JOIN payment_history ph ON ph.employee_id = e.id_employee
		JOIN invoice i ON ph.invoice_id = i.id_invoice
		WHERE i.location_id = ?
		  AND DATE(ph.payment_timestamp) >= ?
		  AND DATE(ph.payment_timestamp) <= ?
		GROUP BY e.id_employee, e.first_name, e.last_name
	`, loc.IDLocation, startDate, endDate).Scan(&rows)

	var data []map[string]interface{}
	for _, r := range rows {
		data = append(data, map[string]interface{}{
			"name":  r.Name,
			"sales": r.Sales,
		})
	}

	return data, nil
}

// ── GetEmployeeInvoices ─────────────────────────────────────────────────────

func (s *Service) GetEmployeeInvoices(username, period string) ([]map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	startDate, endDate := periodDates(period)

	type invRow struct {
		Name            string
		AmountOfInvoice int
	}
	var rows []invRow
	s.db.Raw(`
		SELECT CONCAT(e.first_name, ' ', e.last_name) as name, COUNT(i.id_invoice) as amount_of_invoice
		FROM employee e
		JOIN invoice i ON i.employee_id = e.id_employee
		WHERE i.location_id = ?
		  AND DATE(i.date_create) >= ?
		  AND DATE(i.date_create) <= ?
		GROUP BY e.id_employee, e.first_name, e.last_name
	`, loc.IDLocation, startDate, endDate).Scan(&rows)

	var data []map[string]interface{}
	for _, r := range rows {
		data = append(data, map[string]interface{}{
			"name":              r.Name,
			"amount_of_invoice": r.AmountOfInvoice,
		})
	}

	return data, nil
}

// ── helpers ─────────────────────────────────────────────────────────────────

func periodDates(period string) (string, string) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if period == "d" {
		d := today.Format("2006-01-02")
		return d, d
	}
	// week (default)
	weekday := int(today.Weekday())
	offset := weekday - 1
	if offset < 0 {
		offset = 6
	}
	startOfWeek := today.AddDate(0, 0, -offset)
	endOfWeek := startOfWeek.AddDate(0, 0, 6)
	return startOfWeek.Format("2006-01-02"), endOfWeek.Format("2006-01-02")
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

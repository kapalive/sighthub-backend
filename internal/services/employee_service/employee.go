package employee_service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	auditModel "sighthub-backend/internal/models/audit"
	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	locModel "sighthub-backend/internal/models/location"
	permModel "sighthub-backend/internal/models/permission"
	schedModel "sighthub-backend/internal/models/schedule"
	pkgActivity "sighthub-backend/pkg/activitylog"
	permSvc "sighthub-backend/internal/services/permission_service"
)

type Service struct {
	db      *gorm.DB
	permSvc *permSvc.Service
}

func New(db *gorm.DB) *Service {
	return &Service{db: db, permSvc: permSvc.New(db)}
}

// --- DTOs ---

type EmployeeListItem struct {
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	Username   *string `json:"username"`
	Email      *string `json:"email"`
	ID         int     `json:"id"`
	Doctor     bool    `json:"doctor"`
	JobTitleID *int64  `json:"job_title_id"`
	JobTitle   *string `json:"job_title"`
	Erx        bool    `json:"erx"`
	Active     bool    `json:"active"`
}

type EmployeeGeneralResponse struct {
	Username        *string `json:"username"`
	IDEmployee      int     `json:"id_employee"`
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	PrintingName    *string `json:"printing_name"`
	OutsideVendor   string  `json:"outside_vendor"`
	LocationPayroll *string `json:"location_payroll"`
	PayrollTypeName *string `json:"payroll_type_name"`
	SSN             *string `json:"ssn"`
	Doctor          string  `json:"Doctor"`
	DrNPINumber     *string `json:"dr_npi_number"`
	EINOthLicNum    *string `json:"ein_oth_license_num"`
	Email           *string `json:"email"`
	StartDate       *string `json:"start_date"`
	TerminationDate *string `json:"termination_date"`
	Active          string  `json:"active"`
	JobTitleID      *int64  `json:"job_title_id"`
	JobTitle        *string `json:"job_title"`
	ReferralSig     string  `json:"referral_signature"`
	SignatureImg    *string `json:"signature_img"`
}

type AddEmployeeInput struct {
	General      map[string]interface{} `json:"general"`
	Auth         map[string]interface{} `json:"authentication"`
	Roles        interface{}            `json:"roles"`
	Provider     bool                   `json:"provider"`
	ProviderInfo map[string]interface{} `json:"provider_info"`
	LocationID   int                    `json:"-"`
	RawLocID     interface{}            `json:"location_id"`
	Signature    *string                `json:"signature"`
	JobTitleID   *int64                 `json:"job_title_id"`
}

func (a *AddEmployeeInput) ParseLocationID() {
	a.LocationID = toInt(a.RawLocID)
}

type UpdateEmployeeInput struct {
	General      map[string]interface{} `json:"general"`
	Roles        interface{}            `json:"roles"`
	Provider     bool                   `json:"provider"`
	ProviderInfo map[string]interface{} `json:"provider_info"`
	LocationID   int                    `json:"-"`
	RawLocID     interface{}            `json:"location_id"`
	Signature    *string                `json:"signature"`
	JobTitleID   *int64                 `json:"job_title_id"`
}

// ParseLocationID converts RawLocID (string or number) into LocationID int.
func (u *UpdateEmployeeInput) ParseLocationID() {
	u.LocationID = toInt(u.RawLocID)
}

type EmployeeDetailResponse struct {
	LocationID   *int64                 `json:"location_id"`
	Provider     bool                   `json:"provider"`
	JobTitleID   *int64                 `json:"job_title_id"`
	JobTitle     *string                `json:"job_title"`
	General      map[string]interface{} `json:"general"`
	Signature    *string                `json:"signature"`
	Roles        []string               `json:"roles"`
	ProviderInfo map[string]interface{} `json:"provider_info"`
}

type TimecardListItem struct {
	ID         int     `json:"id"`
	EmployeeID *int    `json:"employee_id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Active     bool    `json:"active"`
	Username   string  `json:"username"`
	LastAction *string `json:"last_action"`
	Timestamp  *string `json:"timestamp"`
}

type TimecardPeriod struct {
	Date     string  `json:"date,omitempty"`
	Checkin  string  `json:"checkin"`
	Checkout string  `json:"checkout"`
	Summary  string  `json:"summary"`
	Note     *string `json:"note"`
}

type TimecardHistoryResponse struct {
	TotalTime string           `json:"total_time"`
	Periods   []TimecardPeriod `json:"periods"`
}

// --- Employee ---

func (s *Service) CheckLoginAvailability(username string) (bool, error) {
	var login authModel.EmployeeLogin
	err := s.db.Where("LOWER(employee_login) = ?", strings.ToLower(username)).First(&login).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func (s *Service) GetLocations() ([]map[string]interface{}, error) {
	var locs []locModel.Location
	s.db.Where("showcase = true AND store_active = true").Find(&locs)
	result := make([]map[string]interface{}, 0, len(locs))
	for _, loc := range locs {
		result = append(result, map[string]interface{}{
			"location_id":   loc.IDLocation,
			"location_name": loc.FullName,
		})
	}
	return result, nil
}

func (s *Service) GetWarehousesByLocation(locationID int) ([]map[string]interface{}, error) {
	var loc locModel.Location
	if err := s.db.First(&loc, locationID).Error; err != nil {
		return nil, errors.New("location not found")
	}
	var whLocs []locModel.Location
	s.db.Where("store_id = ? AND warehouse_id IS NOT NULL", loc.StoreID).Preload("Warehouse").Find(&whLocs)
	result := make([]map[string]interface{}, 0, len(whLocs))
	for _, wl := range whLocs {
		if wl.Warehouse == nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"warehouse_id": wl.WarehouseID,
			"full_name":    wl.Warehouse.FullName,
		})
	}
	return result, nil
}

func (s *Service) GetRoles() ([]map[string]interface{}, error) {
	var roles []permModel.Role
	s.db.Find(&roles)
	result := make([]map[string]interface{}, 0, len(roles))
	for _, r := range roles {
		result = append(result, map[string]interface{}{
			"role_id":   r.RoleID,
			"role_name": r.RoleName,
			"key":       r.Key,
		})
	}
	return result, nil
}

func (s *Service) ListEmployees() ([]EmployeeListItem, error) {
	var employees []empModel.Employee
	s.db.Find(&employees)
	result := make([]EmployeeListItem, 0, len(employees))
	for _, emp := range employees {
		var jt empModel.JobTitle
		isDoctor := false
		var jtID *int64
		var jtTitle *string
		if emp.JobTitleID != nil {
			if s.db.First(&jt, *emp.JobTitleID).Error == nil {
				isDoctor = jt.Doctor
				jtID = emp.JobTitleID
				jtTitle = &jt.Title
			}
		}
		var sig empModel.ElectronReferralSignature
		erx := s.db.Where("employee_id = ?", emp.IDEmployee).First(&sig).Error == nil

		var login authModel.EmployeeLogin
		var username *string
		loginActive := emp.Active
		if s.db.First(&login, emp.EmployeeLoginID).Error == nil {
			username = &login.Username
			loginActive = login.Active
		}
		result = append(result, EmployeeListItem{
			LastName:   emp.LastName,
			FirstName:  emp.FirstName,
			Username:   username,
			Email:      emp.Email,
			ID:         emp.IDEmployee,
			Doctor:     isDoctor,
			JobTitleID: jtID,
			JobTitle:   jtTitle,
			Erx:        erx,
			Active:     loginActive,
		})
	}
	return result, nil
}

func (s *Service) GetEmployeeGeneral(currentUsername string) (*EmployeeGeneralResponse, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", currentUsername).First(&login).Error; err != nil {
		return nil, errors.New("user not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, errors.New("employee not found")
	}

	isDoctor := "No"
	var jtID *int64
	var jtTitle *string
	if emp.JobTitleID != nil {
		var jt empModel.JobTitle
		if s.db.First(&jt, *emp.JobTitleID).Error == nil {
			if jt.Doctor {
				isDoctor = "Yes"
			}
			jtID = emp.JobTitleID
			jtTitle = &jt.Title
		}
	}

	var sig empModel.ElectronReferralSignature
	hasSig := s.db.Where("employee_id = ?", emp.IDEmployee).First(&sig).Error == nil
	refSig := "No"
	if hasSig {
		refSig = "Yes"
	}
	sigImg := sig.PathLinkImg

	resp := &EmployeeGeneralResponse{
		Username:      &login.Username,
		IDEmployee:    emp.IDEmployee,
		FirstName:     emp.FirstName,
		LastName:      emp.LastName,
		OutsideVendor: "No",
		SSN:           emp.SSN,
		Doctor:        isDoctor,
		Email:         emp.Email,
		Active:        "N",
		JobTitleID:    jtID,
		JobTitle:      jtTitle,
		ReferralSig:   refSig,
		SignatureImg:  sigImg,
	}
	if emp.Active {
		resp.Active = "Y"
	}

	// Payroll location
	if emp.StorePayrollID != nil {
		var payLoc locModel.Location
		if s.db.First(&payLoc, *emp.StorePayrollID).Error == nil {
			resp.LocationPayroll = &payLoc.FullName
		}
	}

	// NPI
	var doc empModel.DoctorNpiNumber
	if s.db.Where("employee_id = ?", emp.IDEmployee).First(&doc).Error == nil {
		resp.DrNPINumber = &doc.DRNPINumber
		resp.PrintingName = doc.PrintingName
	}

	if emp.StartDate != nil {
		s := emp.StartDate.Format("2006-01-02")
		resp.StartDate = &s
	}
	if emp.TerminationDate != nil {
		s := emp.TerminationDate.Format("2006-01-02")
		resp.TerminationDate = &s
	}
	return resp, nil
}

func (s *Service) GetEmployee(employeeID int) (*EmployeeDetailResponse, error) {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return nil, errors.New("employee not found")
	}

	var sig empModel.ElectronReferralSignature
	hasSig := s.db.Where("employee_id = ?", emp.IDEmployee).First(&sig).Error == nil

	var roles []empModel.EmployeeRole
	s.db.Where("id_employee_login = ?", emp.EmployeeLoginID).Find(&roles)
	roleNames := make([]string, 0)
	for _, role := range roles {
		var r permModel.Role
		if s.db.First(&r, role.RoleID).Error == nil {
			roleNames = append(roleNames, r.RoleName)
		}
	}

	var doc empModel.DoctorNpiNumber
	hasDoc := s.db.Where("employee_id = ?", emp.IDEmployee).First(&doc).Error == nil
	isProvider := hasDoc && doc.DRNPINumber != ""

	general := map[string]interface{}{
		"first_name":      emp.FirstName,
		"middle_name":     emp.MiddleName,
		"last_name":       emp.LastName,
		"prefix":          emp.Prefix,
		"suffix":          emp.Suffix,
		"dob":             formatDate(emp.DOB),
		"phone":           emp.Phone,
		"email":           emp.Email,
		"street_address":  emp.StreetAddress,
		"address_line_2":  emp.AddressLine2,
		"city":            emp.City,
		"state":           emp.State,
		"zip":             emp.Zip,
		"country":         emp.Country,
		"ssn":             emp.SSN,
		"date_hired":      formatDate(emp.StartDate),
		"date_terminated": formatDate(emp.TerminationDate),
		"active":          emp.Active,
	}

	providerInfo := map[string]interface{}{
		"npi":                   nil,
		"printing_name":         nil,
		"ein":                   nil,
		"dea":                   nil,
		"dea_expiration":        nil,
		"state_license_number":  nil,
		"state_license":         nil,
		"referral_signature":    hasSig,
	}
	if hasDoc {
		providerInfo["npi"] = doc.DRNPINumber
		providerInfo["printing_name"] = doc.PrintingName
		providerInfo["ein"] = doc.EIN
		providerInfo["dea"] = doc.DEA
		if doc.DEAExpiration != nil {
			providerInfo["dea_expiration"] = doc.DEAExpiration.Format("2006-01-02")
		}
	}

	var jtID *int64
	var jtTitle *string
	if emp.JobTitleID != nil {
		var jt empModel.JobTitle
		if s.db.First(&jt, *emp.JobTitleID).Error == nil {
			jtID = emp.JobTitleID
			jtTitle = &jt.Title
		}
	}

	return &EmployeeDetailResponse{
		LocationID:   emp.StorePayrollID,
		Provider:     isProvider,
		JobTitleID:   jtID,
		JobTitle:     jtTitle,
		General:      general,
		Signature:    emp.Signature,
		Roles:        roleNames,
		ProviderInfo: providerInfo,
	}, nil
}

func (s *Service) AddEmployee(currentUsername string, input AddEmployeeInput) (int, error) {
	if input.LocationID == 0 {
		return 0, errors.New("location_id is required")
	}

	authData := input.Auth
	if authData == nil {
		return 0, errors.New("authentication is required")
	}
	username, _ := authData["username"].(string)
	password, _ := authData["password"].(string)
	if username == "" || password == "" {
		return 0, errors.New("username and password are required")
	}

	var existing authModel.EmployeeLogin
	if s.db.Where("LOWER(employee_login) = ?", strings.ToLower(username)).First(&existing).Error == nil {
		return 0, errors.New("username already exists")
	}

	var grantor authModel.EmployeeLogin
	var grantID int
	if s.db.Where("employee_login = ?", currentUsername).First(&grantor).Error == nil {
		grantID = grantor.IDEmployeeLogin
	}

	var empID int
	err := s.db.Transaction(func(tx *gorm.DB) error {
		loginActive := true
		if v, ok := input.General["active"].(bool); ok {
			loginActive = v
		}
		newLogin := &authModel.EmployeeLogin{
			Username: username,
			Active:   loginActive,
		}
		if err := newLogin.SetPassword(password); err != nil {
			return err
		}
		if err := tx.Create(newLogin).Error; err != nil {
			return err
		}

		g := input.General
		newEmp := &empModel.Employee{
			FirstName:       strVal(g, "first_name"),
			LastName:        strVal(g, "last_name"),
			EmployeeLoginID: int64(newLogin.IDEmployeeLogin),
			Active:          loginActive,
			LocationID:      int64Ptr(input.LocationID),
			StorePayrollID:  int64Ptr(input.LocationID),
		}
		newEmp.MiddleName = strPtrVal(g, "middle_name")
		newEmp.Prefix = strPtrVal(g, "prefix")
		newEmp.Suffix = strPtrVal(g, "suffix")
		newEmp.Phone = strPtrVal(g, "phone")
		newEmp.Email = strPtrVal(g, "email")
		newEmp.StreetAddress = strPtrVal(g, "street_address")
		newEmp.AddressLine2 = strPtrVal(g, "address_line_2")
		newEmp.City = strPtrVal(g, "city")
		newEmp.State = strPtrVal(g, "state")
		newEmp.Zip = strPtrVal(g, "zip")
		newEmp.Country = strPtrVal(g, "country")
		newEmp.SSN = strPtrVal(g, "ssn")
		newEmp.DOB = parseDatePtr(g, "dob")
		newEmp.StartDate = parseDatePtr(g, "date_hired")
		if dt := strPtrVal(g, "date_terminated"); dt != nil && *dt != "" {
			newEmp.TerminationDate = parseDatePtr(g, "date_terminated")
		}

		if input.Signature != nil {
			cleaned := cleanSignature(*input.Signature)
			newEmp.Signature = &cleaned
		}

		if input.JobTitleID != nil {
			newEmp.JobTitleID = input.JobTitleID
		} else if input.Provider {
			jid := int64(1)
			newEmp.JobTitleID = &jid
		} else {
			jid := int64(10)
			newEmp.JobTitleID = &jid
		}

		if err := tx.Create(newEmp).Error; err != nil {
			return err
		}
		empID = newEmp.IDEmployee

		// Assign roles
		assignedRoleIDs := s.assignRoles(tx, newLogin.IDEmployeeLogin, input.Roles, grantID)

		// Default permissions
		if len(assignedRoleIDs) > 0 {
			s.permSvc.AssignDefaultPermissionsForRoles(tx, newLogin.IDEmployeeLogin, input.LocationID, assignedRoleIDs, grantID)
		}

		// Provider info
		if input.Provider {
			pi := input.ProviderInfo
			doc := &empModel.DoctorNpiNumber{
				DRNPINumber: strMapVal(pi, "npi"),
				EmployeeID:  int64PtrVal(empID),
			}
			doc.PrintingName = strPtrVal(pi, "printing_name")
			doc.EIN = strPtrVal(pi, "ein")
			doc.DEA = strPtrVal(pi, "dea")
			doc.DEAExpiration = parseDatePtrFromMap(pi, "dea_expiration")
			tx.Create(doc)
		}

		pkgActivity.Log(s.db, "employee", "create",
			pkgActivity.WithEntity(int64(empID)),
			pkgActivity.WithDetails(map[string]interface{}{
				"name":        newEmp.FirstName + " " + newEmp.LastName,
				"location_id": input.LocationID,
			}),
		)
		return nil
	})
	return empID, err
}

func (s *Service) UpdateEmployee(currentUsername string, employeeID int, input UpdateEmployeeInput) error {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return errors.New("employee not found")
	}

	var grantor authModel.EmployeeLogin
	var granterID int
	if s.db.Where("employee_login = ?", currentUsername).First(&grantor).Error == nil {
		granterID = grantor.IDEmployeeLogin
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		g := input.General
		if v := strMapVal(g, "first_name"); v != "" {
			emp.FirstName = v
		}
		if v := strMapVal(g, "last_name"); v != "" {
			emp.LastName = v
		}
		emp.MiddleName = strPtrVal(g, "middle_name")
		emp.Prefix = strPtrVal(g, "prefix")
		emp.Suffix = strPtrVal(g, "suffix")
		emp.Phone = strPtrVal(g, "phone")
		emp.Email = strPtrVal(g, "email")
		emp.StreetAddress = strPtrVal(g, "street_address")
		emp.AddressLine2 = strPtrVal(g, "address_line_2")
		emp.City = strPtrVal(g, "city")
		emp.State = strPtrVal(g, "state")
		emp.Zip = strPtrVal(g, "zip")
		emp.Country = strPtrVal(g, "country")
		emp.SSN = strPtrVal(g, "ssn")
		if dob := parseDatePtr(g, "dob"); dob != nil {
			emp.DOB = dob
		}
		if sd := parseDatePtr(g, "date_hired"); sd != nil {
			emp.StartDate = sd
		}
		if v, ok := g["date_terminated"]; ok {
			if v == nil || v == "" {
				emp.TerminationDate = nil
			} else {
				emp.TerminationDate = parseDatePtr(g, "date_terminated")
			}
		}
		if v, ok := g["active"].(bool); ok {
			emp.Active = v
			var login authModel.EmployeeLogin
			if tx.First(&login, emp.EmployeeLoginID).Error == nil {
				login.Active = v
				tx.Save(&login)
			}
		}

		if input.Signature != nil {
			cleaned := cleanSignature(*input.Signature)
			emp.Signature = &cleaned
		}

		if input.LocationID != 0 {
			lid := int64(input.LocationID)
			emp.LocationID = &lid
			emp.StorePayrollID = &lid
		}

		if input.JobTitleID != nil {
			emp.JobTitleID = input.JobTitleID
		} else if input.Provider {
			jid := int64(1)
			emp.JobTitleID = &jid
		} else {
			jid := int64(10)
			emp.JobTitleID = &jid
		}

		if err := tx.Save(&emp).Error; err != nil {
			return err
		}

		// Roles
		tx.Where("id_employee_login = ?", emp.EmployeeLoginID).Delete(&empModel.EmployeeRole{})
		assignedRoleIDs := s.assignRoles(tx, int(emp.EmployeeLoginID), input.Roles, 0)
		if len(assignedRoleIDs) > 0 && input.LocationID != 0 {
			s.permSvc.AssignDefaultPermissionsForRoles(tx, int(emp.EmployeeLoginID), input.LocationID, assignedRoleIDs, granterID)
		}

		// Provider
		var doc empModel.DoctorNpiNumber
		hasDoc := tx.Where("employee_id = ?", emp.IDEmployee).First(&doc).Error == nil
		if input.Provider {
			pi := input.ProviderInfo
			if hasDoc {
				doc.DRNPINumber = strMapVal(pi, "npi")
				doc.PrintingName = strPtrVal(pi, "printing_name")
				doc.EIN = strPtrVal(pi, "ein")
				doc.DEA = strPtrVal(pi, "dea")
				doc.DEAExpiration = parseDatePtrFromMap(pi, "dea_expiration")
				tx.Save(&doc)
			} else {
				newDoc := &empModel.DoctorNpiNumber{
					DRNPINumber: strMapVal(pi, "npi"),
					EmployeeID:  int64PtrVal(emp.IDEmployee),
				}
				newDoc.PrintingName = strPtrVal(pi, "printing_name")
				newDoc.EIN = strPtrVal(pi, "ein")
				newDoc.DEA = strPtrVal(pi, "dea")
				newDoc.DEAExpiration = parseDatePtrFromMap(pi, "dea_expiration")
				tx.Create(newDoc)
			}
		} else if hasDoc {
			tx.Delete(&doc)
		}

		pkgActivity.Log(s.db, "employee", "update",
			pkgActivity.WithEntity(int64(employeeID)),
			pkgActivity.WithDetails(map[string]interface{}{"updated_fields": getKeys(input.General)}),
		)
		return nil
	})
}

// --- Timecard ---

func (s *Service) ListTimecards(activeOnly bool) ([]TimecardListItem, error) {
	var timecards []empModel.EmployeeTimecardLogin
	query := s.db
	if activeOnly {
		query = query.Where("active = true")
	}
	query.Find(&timecards)

	result := make([]TimecardListItem, 0, len(timecards))
	for _, tc := range timecards {
		var last auditModel.EmployeeTimecardHistory
		hasLast := s.db.Where("employee_timecard_login_id = ?", tc.IDEmployeeTimecardLogin).
			Order("timestamp DESC").First(&last).Error == nil

		item := TimecardListItem{
			ID:         tc.IDEmployeeTimecardLogin,
			EmployeeID: tc.EmployeeID,
			FirstName:  tc.FirstName,
			LastName:   tc.LastName,
			Active:     tc.Active,
			Username:   tc.Username,
		}
		if hasLast {
			item.LastAction = &last.ActionType
			ts := last.Timestamp.Format(time.RFC3339)
			item.Timestamp = &ts
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) GetTimecardHistory(timecardLoginID int, startDate, endDate time.Time, withDate bool) (*TimecardHistoryResponse, error) {
	var tc empModel.EmployeeTimecardLogin
	if err := s.db.First(&tc, timecardLoginID).Error; err != nil {
		return nil, errors.New("timecard account not found")
	}

	// Check timecard permission if linked to employee
	if tc.EmployeeID != nil {
		var emp empModel.Employee
		if s.db.First(&emp, *tc.EmployeeID).Error == nil {
			hasPerm := s.db.Table("employee_permission ep").
				Joins("JOIN permissions_combination pc ON pc.id_permissions_combination = ep.permissions_combination_id").
				Where("ep.employee_login_id = ? AND ep.is_active = true AND pc.permissions_id = 50", emp.EmployeeLoginID).
				First(&empModel.EmployeePermission{}).Error == nil
			if !hasPerm {
				return nil, errors.New("timecard permission disabled for this employee")
			}
		}
	}

	return s.buildTimecardHistory(timecardLoginID, startDate, endDate, withDate)
}

func (s *Service) GetEmployeeTimecardHistory(employeeID int, startDate, endDate time.Time) (*TimecardHistoryResponse, error) {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	var tc empModel.EmployeeTimecardLogin
	if err := s.db.Where("employee_id = ?", employeeID).First(&tc).Error; err != nil {
		return nil, errors.New("timecard account not found")
	}
	hasPerm := s.db.Table("employee_permission ep").
		Joins("JOIN permissions_combination pc ON pc.id_permissions_combination = ep.permissions_combination_id").
		Where("ep.employee_login_id = ? AND ep.is_active = true AND pc.permissions_id = 50", emp.EmployeeLoginID).
		First(&empModel.EmployeePermission{}).Error == nil
	if !hasPerm {
		return nil, errors.New("timecard permission disabled for this employee")
	}
	return s.buildTimecardHistory(tc.IDEmployeeTimecardLogin, startDate, endDate, false)
}

func (s *Service) CreateTimecard(employeeID *int, username, password, firstName, lastName string) (int, error) {
	if s.db.Where("username = ?", username).First(&empModel.EmployeeTimecardLogin{}).Error == nil {
		return 0, errors.New("username already exists")
	}
	if employeeID != nil {
		if s.db.Where("employee_id = ?", *employeeID).First(&empModel.EmployeeTimecardLogin{}).Error == nil {
			return 0, errors.New("timecard account already exists")
		}
		if firstName == "" || lastName == "" {
			var emp empModel.Employee
			if s.db.First(&emp, *employeeID).Error == nil {
				firstName = emp.FirstName
				lastName = emp.LastName
			}
		}
	}
	tc := &empModel.EmployeeTimecardLogin{
		EmployeeID: employeeID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
		Active:     true,
	}
	if err := tc.SetPassword(password); err != nil {
		return 0, err
	}
	if err := s.db.Create(tc).Error; err != nil {
		return 0, err
	}
	return tc.IDEmployeeTimecardLogin, nil
}

func (s *Service) UpdateTimecard(employeeID int, username, password, firstName, lastName string) error {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return errors.New("employee not found")
	}
	var tc empModel.EmployeeTimecardLogin
	if err := s.db.Where("employee_id = ?", employeeID).First(&tc).Error; err != nil {
		return errors.New("timecard account not found")
	}
	if username != "" {
		var existing empModel.EmployeeTimecardLogin
		if s.db.Where("username = ?", username).First(&existing).Error == nil &&
			existing.IDEmployeeTimecardLogin != tc.IDEmployeeTimecardLogin {
			return errors.New("username already exists")
		}
		tc.Username = username
	}
	if password != "" {
		if err := tc.SetPassword(password); err != nil {
			return err
		}
	}
	if firstName != "" {
		tc.FirstName = firstName
	}
	if lastName != "" {
		tc.LastName = lastName
	}
	return s.db.Save(&tc).Error
}

func (s *Service) DeactivateTimecard(loginID int) error {
	var tc empModel.EmployeeTimecardLogin
	if err := s.db.First(&tc, loginID).Error; err != nil {
		return errors.New("timecard user not found")
	}
	tc.Active = false
	pkgActivity.Log(s.db, "employee", "timecard_deactivate", pkgActivity.WithEntity(int64(loginID)))
	return s.db.Save(&tc).Error
}

// --- Schedule ---

func (s *Service) GetSchedule(employeeID int, startDate, endDate *time.Time) (interface{}, error) {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	if startDate == nil {
		return buildWeeklySchedule(s.db, employeeID), nil
	}
	return buildScheduleRange(s.db, employeeID, *startDate, *endDate), nil
}

func (s *Service) CreateSchedule(employeeID int, data map[string]interface{}) error {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return errors.New("employee not found")
	}
	saveWeeklySchedule(s.db, employeeID, data)
	pkgActivity.Log(s.db, "employee", "schedule_create", pkgActivity.WithEntity(int64(employeeID)))
	return nil
}

func (s *Service) UpdateSchedule(employeeID int, data map[string]interface{}) (interface{}, error) {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	saveWeeklySchedule(s.db, employeeID, data)
	pkgActivity.Log(s.db, "employee", "schedule_update", pkgActivity.WithEntity(int64(employeeID)))
	return buildWeeklySchedule(s.db, employeeID), nil
}

// --- Off days ---

func (s *Service) AddOffDay(employeeID int, dateStr string) (string, error) {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return "", errors.New("employee not found")
	}
	target, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", errors.New("invalid date format. Use YYYY-MM-DD")
	}
	var cal schedModel.Calendar
	empID64 := int64(employeeID)
	if s.db.Where("employee_id = ? AND date = ?", employeeID, target).First(&cal).Error == nil {
		cal.IsWorkingDay = false
		cal.IsHoliday = true
		s.db.Save(&cal)
	} else {
		s.db.Create(&schedModel.Calendar{
			Date:         target,
			IsHoliday:    true,
			IsWorkingDay: false,
			EmployeeID:   &empID64,
			WorkShiftID:  emp.WorkShiftID,
		})
	}
	return target.Format("2006-01-02"), nil
}

func (s *Service) ListOffDays(employeeID int) ([]string, error) {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	var entries []schedModel.Calendar
	s.db.Where("employee_id = ? AND is_working_day = false", employeeID).Find(&entries)
	result := make([]string, 0, len(entries))
	for _, e := range entries {
		result = append(result, e.Date.Format("2006-01-02"))
	}
	return result, nil
}

func (s *Service) RemoveOffDay(employeeID int, dateStr string) (string, error) {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return "", errors.New("employee not found")
	}
	target, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", errors.New("invalid date format. Use YYYY-MM-DD")
	}
	var cal schedModel.Calendar
	if err := s.db.Where("employee_id = ? AND date = ?", employeeID, target).First(&cal).Error; err != nil || cal.IsWorkingDay {
		return "", errors.New("off day not found")
	}
	dayOfWeek := strings.ToLower(target.Weekday().String())
	var sched schedModel.Schedule
	if s.db.Where("employee_id = ? AND day_of_week = ?", employeeID, dayOfWeek).First(&sched).Error == nil {
		cal.IsWorkingDay = true
		cal.IsHoliday = false
		s.db.Save(&cal)
	} else {
		s.db.Delete(&cal)
	}
	return target.Format("2006-01-02"), nil
}

// --- Job Titles ---

func (s *Service) GetJobTitles() ([]map[string]interface{}, error) {
	var titles []empModel.JobTitle
	if err := s.db.Order("title").Find(&titles).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(titles))
	for _, t := range titles {
		result = append(result, t.ToMap())
	}
	return result, nil
}

// --- Helpers ---

func (s *Service) assignRoles(tx *gorm.DB, loginID int, roles interface{}, grantID int) []int {
	var grantIDPtr *int
	if grantID != 0 {
		grantIDPtr = &grantID
	}
	roleItems := parseRoleItems(roles)
	assigned := []int{}
	for key, flag := range roleItems {
		if !flag {
			continue
		}
		var role permModel.Role
		if tx.Where("key = ? OR role_name = ?", key, key).First(&role).Error != nil {
			continue
		}
		tx.Create(&empModel.EmployeeRole{
			IDEmployeeLogin: loginID,
			RoleID:          role.RoleID,
			GrantedBy:       grantIDPtr,
		})
		assigned = append(assigned, role.RoleID)
	}
	return assigned
}

func parseRoleItems(roles interface{}) map[string]bool {
	result := map[string]bool{}
	if roles == nil {
		return result
	}
	switch v := roles.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if b, ok := val.(bool); ok {
				result[key] = b
			}
		}
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				key := ""
				if k, ok := m["key"].(string); ok {
					key = k
				} else if k, ok := m["name"].(string); ok {
					key = k
				}
				if key != "" {
					if b, ok := m[key].(bool); ok {
						result[key] = b
					}
				}
			}
		}
	}
	return result
}

func buildWeeklySchedule(db *gorm.DB, employeeID int) map[string]interface{} {
	days := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	result := map[string]interface{}{}
	for _, day := range days {
		var sched schedModel.Schedule
		if db.Where("employee_id = ? AND day_of_week = ?", employeeID, day).First(&sched).Error == nil {
			result[day] = map[string]interface{}{
				"is_working":           true,
				"start_time":           sched.StartTime,
				"end_time":             sched.EndTime,
				"appointment_duration": sched.AppointmentDuration,
			}
		} else {
			result[day] = map[string]interface{}{"is_working": false}
		}
	}
	return result
}

func buildScheduleRange(db *gorm.DB, employeeID int, start, end time.Time) []map[string]interface{} {
	result := []map[string]interface{}{}
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dayOfWeek := strings.ToLower(d.Weekday().String())
		dateStr := d.Format("2006-01-02")

		var cal schedModel.Calendar
		isOffDay := db.Where("employee_id = ? AND date = ? AND is_working_day = false", employeeID, d).First(&cal).Error == nil

		if isOffDay {
			result = append(result, map[string]interface{}{
				"date": dateStr, "is_working_day": false, "message": "Off day",
				"time_start": nil, "time_end": nil,
			})
			continue
		}

		var sched schedModel.Schedule
		if db.Where("employee_id = ? AND day_of_week = ?", employeeID, dayOfWeek).First(&sched).Error == nil {
			result = append(result, map[string]interface{}{
				"date": dateStr, "is_working_day": true,
				"time_start":           sched.StartTime,
				"time_end":             sched.EndTime,
				"appointment_duration": sched.AppointmentDuration,
			})
		} else {
			result = append(result, map[string]interface{}{
				"date": dateStr, "is_working_day": false, "message": "Not scheduled",
				"time_start": nil, "time_end": nil,
			})
		}
	}
	return result
}

func saveWeeklySchedule(db *gorm.DB, employeeID int, data map[string]interface{}) {
	days := []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}
	for _, day := range days {
		info, _ := data[day].(map[string]interface{})
		var sched schedModel.Schedule
		exists := db.Where("employee_id = ? AND day_of_week = ?", employeeID, day).First(&sched).Error == nil

		if info == nil || !boolVal(info, "is_working", true) {
			if exists {
				db.Delete(&sched)
			}
			continue
		}
		startStr, _ := info["start_time"].(string)
		endStr, _ := info["end_time"].(string)
		if startStr == "" || endStr == "" {
			continue
		}
		startT, err1 := time.Parse("15:04", startStr)
		endT, err2 := time.Parse("15:04", endStr)
		if err1 != nil || err2 != nil {
			continue
		}
		dur := 15
		if d, ok := info["appointment_duration"].(float64); ok {
			dur = int(d)
		}
		if exists {
			sched.StartTime = startT.Format("15:04:05")
			sched.EndTime = endT.Format("15:04:05")
			sched.AppointmentDuration = dur
			db.Save(&sched)
		} else {
			db.Create(&schedModel.Schedule{
				EmployeeID:          int64(employeeID),
				DayOfWeek:           day,
				StartTime:           startT.Format("15:04:05"),
				EndTime:             endT.Format("15:04:05"),
				AppointmentDuration: dur,
			})
		}
	}
}

func (s *Service) buildTimecardHistory(tcID int, start, end time.Time, withDate bool) (*TimecardHistoryResponse, error) {
	var history []auditModel.EmployeeTimecardHistory
	s.db.Where("employee_timecard_login_id = ? AND timestamp >= ? AND timestamp <= ?", tcID, start, end).
		Order("timestamp").Find(&history)

	var periods []TimecardPeriod
	var totalSeconds float64
	var lastCheckin *auditModel.EmployeeTimecardHistory
	for i := range history {
		entry := &history[i]
		if entry.ActionType == "checkin" {
			lastCheckin = entry
		} else if entry.ActionType == "checkout" && lastCheckin != nil {
			diff := entry.Timestamp.Sub(lastCheckin.Timestamp)
			totalSeconds += diff.Seconds()
			sec := int(diff.Seconds())
			p := TimecardPeriod{
				Checkin:  lastCheckin.Timestamp.Format("15:04:05 01/02/2006"),
				Checkout: entry.Timestamp.Format("15:04:05 01/02/2006"),
				Summary:  formatDuration(sec),
				Note:     lastCheckin.Note,
			}
			if withDate {
				p.Date = lastCheckin.Timestamp.Format("01/02/2006")
			}
			periods = append(periods, p)
			lastCheckin = nil
		}
	}
	// Reverse
	for i, j := 0, len(periods)-1; i < j; i, j = i+1, j-1 {
		periods[i], periods[j] = periods[j], periods[i]
	}
	if periods == nil {
		periods = []TimecardPeriod{}
	}
	return &TimecardHistoryResponse{
		TotalTime: formatDuration(int(totalSeconds)),
		Periods:   periods,
	}, nil
}

func formatDuration(sec int) string {
	h := sec / 3600
	m := (sec % 3600) / 60
	return fmt.Sprintf("%d:%02d", h, m)
}

// --- Pure helpers ---

func cleanSignature(raw string) string {
	if strings.Contains(raw, "/mnt/tank/data/") {
		parts := strings.SplitN(raw, "/mnt/tank/data/", 2)
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(raw)
}

func strVal(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func strMapVal(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func strPtrVal(m map[string]interface{}, key string) *string {
	if m == nil {
		return nil
	}
	v, ok := m[key].(string)
	if !ok || v == "" {
		return nil
	}
	return &v
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case string:
		n := 0
		fmt.Sscanf(val, "%d", &n)
		return n
	case int:
		return val
	case int64:
		return int(val)
	}
	return 0
}

func int64Ptr(id int) *int64 {
	v := int64(id)
	return &v
}

func int64PtrVal(id int) *int64 {
	v := int64(id)
	return &v
}

func parseDatePtr(m map[string]interface{}, key string) *time.Time {
	s, ok := m[key].(string)
	if !ok || s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

func parseDatePtrFromMap(m map[string]interface{}, key string) *time.Time {
	return parseDatePtr(m, key)
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func boolVal(m map[string]interface{}, key string, def bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return def
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

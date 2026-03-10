package employees_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/auth"
	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/permission"
)

// ─────────────────────────────────────────────
// DTO types
// ─────────────────────────────────────────────

type EmployeeListItem struct {
	IDEmployee int     `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Username   string  `json:"username"`
	Email      *string `json:"email"`
	Doctor     bool    `json:"doctor"`
	ERX        bool    `json:"erx"`
	Active     bool    `json:"active"`
}

type EmployeeDetail struct {
	General      map[string]interface{} `json:"general"`
	ProviderInfo map[string]interface{} `json:"provider_info"`
	Roles        []string               `json:"roles"`
	HasSignature bool                   `json:"has_signature"`
}

type CreateEmployeeInput struct {
	FirstName     string  `json:"first_name"`
	MiddleName    *string `json:"middle_name"`
	LastName      string  `json:"last_name"`
	Suffix        *string `json:"suffix"`
	DOB           *string `json:"dob"`
	Phone         *string `json:"phone"`
	Email         *string `json:"email"`
	StreetAddress *string `json:"street_address"`
	AddressLine2  *string `json:"address_line_2"`
	City          *string `json:"city"`
	State         *string `json:"state"`
	SSN           *string `json:"ssn"`
	StartDate     *string `json:"start_date"`
	Prefix        *string `json:"prefix"`
	Zip           *string `json:"zip"`
	Country       *string `json:"country"`

	Username string `json:"username"`
	Password string `json:"password"`

	RoleIDs []int `json:"role_ids"`

	IsProvider   bool    `json:"is_provider"`
	NPI          *string `json:"npi"`
	EIN          *string `json:"ein"`
	DEA          *string `json:"dea"`
	PrintingName *string `json:"printing_name"`

	GrantedByLoginID int `json:"granted_by_login_id"`
	StoreID          int `json:"store_id"`
}

type UpdateEmployeeInput struct {
	FirstName       *string `json:"first_name"`
	MiddleName      *string `json:"middle_name"`
	LastName        *string `json:"last_name"`
	Suffix          *string `json:"suffix"`
	DOB             *string `json:"dob"`
	Phone           *string `json:"phone"`
	Email           *string `json:"email"`
	StreetAddress   *string `json:"street_address"`
	AddressLine2    *string `json:"address_line_2"`
	City            *string `json:"city"`
	State           *string `json:"state"`
	SSN             *string `json:"ssn"`
	StartDate       *string `json:"start_date"`
	TerminationDate *string `json:"termination_date"`
	Prefix          *string `json:"prefix"`
	Zip             *string `json:"zip"`
	Country         *string `json:"country"`
	Active          *bool   `json:"active"`

	NewPassword *string `json:"new_password"`

	RoleIDs []int `json:"role_ids"`

	IsProvider   *bool   `json:"is_provider"`
	NPI          *string `json:"npi"`
	EIN          *string `json:"ein"`
	DEA          *string `json:"dea"`
	PrintingName *string `json:"printing_name"`

	GrantedByLoginID int `json:"granted_by_login_id"`
	StoreID          int `json:"store_id"`
}

// ─────────────────────────────────────────────
// EmployeeRepo
// ─────────────────────────────────────────────

type EmployeeRepo struct {
	DB *gorm.DB
}

func NewEmployeeRepo(db *gorm.DB) *EmployeeRepo {
	return &EmployeeRepo{DB: db}
}

// GetList returns a lightweight list of all employees.
func (r *EmployeeRepo) GetList() ([]EmployeeListItem, error) {
	type listRow struct {
		IDEmployee int
		FirstName  string
		LastName   string
		Username   string
		Email      *string
		Doctor     bool
		ERXCount   int64
		Active     bool
	}

	var rows []listRow
	err := r.DB.
		Table("employee e").
		Select(`e.id_employee,
			e.first_name,
			e.last_name,
			el.employee_login  AS username,
			e.email,
			COALESCE(jt.doctor, false) AS doctor,
			(SELECT COUNT(*) FROM electron_referral_signature ers WHERE ers.employee_id = e.id_employee) AS erx_count,
			el.active`).
		Joins("JOIN employee_login el ON el.id_employee_login = e.employee_login_id").
		Joins("LEFT JOIN job_title jt ON jt.id_job_title = e.job_title_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]EmployeeListItem, 0, len(rows))
	for _, rw := range rows {
		items = append(items, EmployeeListItem{
			IDEmployee: rw.IDEmployee,
			FirstName:  rw.FirstName,
			LastName:   rw.LastName,
			Username:   rw.Username,
			Email:      rw.Email,
			Doctor:     rw.Doctor,
			ERX:        rw.ERXCount > 0,
			Active:     rw.Active,
		})
	}
	return items, nil
}

// GetByID returns the full employee detail for a given employee ID.
func (r *EmployeeRepo) GetByID(employeeID int) (*EmployeeDetail, error) {
	var emp employees.Employee
	if err := r.DB.First(&emp, "id_employee = ?", employeeID).Error; err != nil {
		return nil, err
	}

	var login auth.EmployeeLogin
	if err := r.DB.First(&login, "id_employee_login = ?", emp.EmployeeLoginID).Error; err != nil {
		return nil, err
	}

	general := emp.ToMap()
	general["username"] = login.Username
	general["employee_login_id"] = emp.EmployeeLoginID

	// Provider info (nil when not a provider)
	var providerInfo map[string]interface{}
	var npi employees.DoctorNpiNumber
	if err := r.DB.Where("employee_id = ?", employeeID).First(&npi).Error; err == nil {
		providerInfo = map[string]interface{}{
			"npi":          npi.DRNPINumber,
			"ein":          npi.EIN,
			"dea":          npi.DEA,
			"printing_name": npi.PrintingName,
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Signature existence check
	var sigCount int64
	r.DB.Model(&employees.ElectronReferralSignature{}).
		Where("employee_id = ?", employeeID).
		Count(&sigCount)

	// Role names
	type roleRow struct {
		RoleName string
	}
	var roleRows []roleRow
	r.DB.Table("employee_roles er").
		Select("r.role_name").
		Joins("JOIN roles r ON r.role_id = er.role_id").
		Where("er.id_employee_login = ?", emp.EmployeeLoginID).
		Scan(&roleRows)

	roleNames := make([]string, 0, len(roleRows))
	for _, rr := range roleRows {
		roleNames = append(roleNames, rr.RoleName)
	}

	return &EmployeeDetail{
		General:      general,
		ProviderInfo: providerInfo,
		Roles:        roleNames,
		HasSignature: sigCount > 0,
	}, nil
}

// CheckLoginAvailability returns true when the username is not yet taken (case-insensitive).
func (r *EmployeeRepo) CheckLoginAvailability(username string) (bool, error) {
	var count int64
	err := r.DB.Model(&auth.EmployeeLogin{}).
		Where("LOWER(employee_login) = LOWER(?)", username).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// Create creates all required records for a new employee inside a single transaction.
// Returns the new id_employee on success.
func (r *EmployeeRepo) Create(input CreateEmployeeInput) (int, error) {
	var newEmployeeID int

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Create EmployeeLogin
		login, err := auth.NewEmployeeLogin(input.Username, input.Password, true)
		if err != nil {
			return err
		}
		if err := tx.Create(login).Error; err != nil {
			return err
		}

		// 2. Determine job_title_id: 1 = provider, 10 = staff
		jobTitleID := int64(10)
		if input.IsProvider {
			jobTitleID = 1
		}

		// 3. Parse optional dates
		var dob *time.Time
		if input.DOB != nil {
			t, err := time.Parse("2006-01-02", *input.DOB)
			if err != nil {
				return err
			}
			dob = &t
		}
		var startDate *time.Time
		if input.StartDate != nil {
			t, err := time.Parse("2006-01-02", *input.StartDate)
			if err != nil {
				return err
			}
			startDate = &t
		}

		// 4. Create Employee
		emp := &employees.Employee{
			FirstName:       input.FirstName,
			MiddleName:      input.MiddleName,
			LastName:        input.LastName,
			Suffix:          input.Suffix,
			DOB:             dob,
			Phone:           input.Phone,
			Email:           input.Email,
			StreetAddress:   input.StreetAddress,
			AddressLine2:    input.AddressLine2,
			City:            input.City,
			State:           input.State,
			SSN:             input.SSN,
			StartDate:       startDate,
			Prefix:          input.Prefix,
			Zip:             input.Zip,
			Country:         input.Country,
			Active:          true,
			EmployeeLoginID: int64(login.IDEmployeeLogin),
			JobTitleID:      &jobTitleID,
		}
		if err := tx.Create(emp).Error; err != nil {
			return err
		}
		newEmployeeID = emp.IDEmployee

		// 5. Create EmployeeRoles
		rolesRepo := NewEmployeeRolesRepo(tx)
		if err := rolesRepo.SetRoles(tx, login.IDEmployeeLogin, input.GrantedByLoginID, input.RoleIDs); err != nil {
			return err
		}

		// 6. Create DoctorNpiNumber if provider
		if input.IsProvider && input.NPI != nil {
			empID64 := int64(emp.IDEmployee)
			npi := &employees.DoctorNpiNumber{
				DRNPINumber:  *input.NPI,
				EIN:          input.EIN,
				DEA:          input.DEA,
				PrintingName: input.PrintingName,
				EmployeeID:   &empID64,
			}
			if err := tx.Create(npi).Error; err != nil {
				return err
			}
		}

		// 7. Assign default permissions for the given roles
		return assignDefaultPermissionsForRoles(tx, login.IDEmployeeLogin, input.StoreID, input.RoleIDs, input.GrantedByLoginID)
	})
	if err != nil {
		return 0, err
	}
	return newEmployeeID, nil
}

// Update updates employee, login, roles and provider info in a single transaction.
func (r *EmployeeRepo) Update(employeeID int, input UpdateEmployeeInput) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var emp employees.Employee
		if err := tx.First(&emp, "id_employee = ?", employeeID).Error; err != nil {
			return err
		}

		empUpdates := map[string]interface{}{}
		if input.FirstName != nil {
			empUpdates["first_name"] = *input.FirstName
		}
		if input.MiddleName != nil {
			empUpdates["middle_name"] = *input.MiddleName
		}
		if input.LastName != nil {
			empUpdates["last_name"] = *input.LastName
		}
		if input.Suffix != nil {
			empUpdates["suffix"] = *input.Suffix
		}
		if input.Phone != nil {
			empUpdates["phone"] = *input.Phone
		}
		if input.Email != nil {
			empUpdates["email"] = *input.Email
		}
		if input.StreetAddress != nil {
			empUpdates["street_address"] = *input.StreetAddress
		}
		if input.AddressLine2 != nil {
			empUpdates["address_line_2"] = *input.AddressLine2
		}
		if input.City != nil {
			empUpdates["city"] = *input.City
		}
		if input.State != nil {
			empUpdates["state"] = *input.State
		}
		if input.SSN != nil {
			empUpdates["ssn"] = *input.SSN
		}
		if input.Prefix != nil {
			empUpdates["prefix"] = *input.Prefix
		}
		if input.Zip != nil {
			empUpdates["zip"] = *input.Zip
		}
		if input.Country != nil {
			empUpdates["country"] = *input.Country
		}
		if input.Active != nil {
			empUpdates["active"] = *input.Active
		}
		if input.DOB != nil {
			t, err := time.Parse("2006-01-02", *input.DOB)
			if err != nil {
				return err
			}
			empUpdates["dob"] = t
		}
		if input.StartDate != nil {
			t, err := time.Parse("2006-01-02", *input.StartDate)
			if err != nil {
				return err
			}
			empUpdates["start_date"] = t
		}
		if input.TerminationDate != nil {
			t, err := time.Parse("2006-01-02", *input.TerminationDate)
			if err != nil {
				return err
			}
			empUpdates["termination_date"] = t
		}

		if len(empUpdates) > 0 {
			if err := tx.Model(&emp).Updates(empUpdates).Error; err != nil {
				return err
			}
		}

		// Update EmployeeLogin
		var login auth.EmployeeLogin
		if err := tx.First(&login, "id_employee_login = ?", emp.EmployeeLoginID).Error; err != nil {
			return err
		}
		loginUpdates := map[string]interface{}{}
		if input.NewPassword != nil {
			if err := login.SetPassword(*input.NewPassword); err != nil {
				return err
			}
			loginUpdates["password_hash"] = login.PasswordHash
		}
		if input.Active != nil {
			loginUpdates["active"] = *input.Active
		}
		if len(loginUpdates) > 0 {
			if err := tx.Model(&login).Updates(loginUpdates).Error; err != nil {
				return err
			}
		}

		// Update roles if caller passed a non-nil slice
		if input.RoleIDs != nil {
			rolesRepo := NewEmployeeRolesRepo(tx)
			if err := rolesRepo.SetRoles(tx, int(emp.EmployeeLoginID), input.GrantedByLoginID, input.RoleIDs); err != nil {
				return err
			}
			if err := assignDefaultPermissionsForRoles(tx, int(emp.EmployeeLoginID), input.StoreID, input.RoleIDs, input.GrantedByLoginID); err != nil {
				return err
			}
		}

		// Update provider info
		if input.IsProvider != nil {
			if *input.IsProvider {
				empID64 := int64(employeeID)
				var npi employees.DoctorNpiNumber
				err := tx.Where("employee_id = ?", employeeID).First(&npi).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					npiVal := ""
					if input.NPI != nil {
						npiVal = *input.NPI
					}
					npi = employees.DoctorNpiNumber{
						DRNPINumber:  npiVal,
						EIN:          input.EIN,
						DEA:          input.DEA,
						PrintingName: input.PrintingName,
						EmployeeID:   &empID64,
					}
					if err := tx.Create(&npi).Error; err != nil {
						return err
					}
				} else if err != nil {
					return err
				} else {
					updates := map[string]interface{}{}
					if input.NPI != nil {
						updates["dr_npi_number"] = *input.NPI
					}
					if input.EIN != nil {
						updates["ein"] = *input.EIN
					}
					if input.DEA != nil {
						updates["dea"] = *input.DEA
					}
					if input.PrintingName != nil {
						updates["printing_name"] = *input.PrintingName
					}
					if len(updates) > 0 {
						if err := tx.Model(&npi).Updates(updates).Error; err != nil {
							return err
						}
					}
				}
			} else {
				tx.Where("employee_id = ?", employeeID).Delete(&employees.DoctorNpiNumber{})
			}
		}

		return nil
	})
}

// GetForLoginID returns the Employee and EmployeeLogin for the given employee_login_id.
func (r *EmployeeRepo) GetForLoginID(loginID int) (*employees.Employee, *auth.EmployeeLogin, error) {
	var login auth.EmployeeLogin
	if err := r.DB.First(&login, "id_employee_login = ?", loginID).Error; err != nil {
		return nil, nil, err
	}

	var emp employees.Employee
	if err := r.DB.First(&emp, "employee_login_id = ?", loginID).Error; err != nil {
		return nil, nil, err
	}

	return &emp, &login, nil
}

// ─────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────

// assignDefaultPermissionsForRoles assigns EmployeePermission entries based on roles.
//
// Logic mirrors Python assign_default_permissions_for_roles():
//  1. Find PermissionsSubBlockStore for the given store (may not exist).
//  2. Collect distinct permissions_ids from RolePermissions for all roleIDs.
//  3. Find PermissionsCombinations with matching permissions_id and either
//     no sub_block_store_id (general) or the store's sub_block id.
//  4. For each combination, upsert EmployeePermission (create if absent, reactivate if inactive).
func assignDefaultPermissionsForRoles(tx *gorm.DB, employeeLoginID, storeID int, roleIDs []int, grantedBy int) error {
	if len(roleIDs) == 0 {
		return nil
	}

	// Resolve store sub-block (best-effort)
	var storeSubBlock permission.PermissionsSubBlockStore
	hasSubBlock := false
	if err := tx.Where("store_id = ?", storeID).First(&storeSubBlock).Error; err == nil {
		hasSubBlock = true
	}

	// Collect distinct permissions_ids from role_permissions
	var rolePerms []permission.RolePermission
	if err := tx.Where("role_id IN ?", roleIDs).Find(&rolePerms).Error; err != nil {
		return err
	}
	permIDSet := make(map[int]struct{}, len(rolePerms))
	for _, rp := range rolePerms {
		permIDSet[rp.PermissionsID] = struct{}{}
	}
	if len(permIDSet) == 0 {
		return nil
	}
	permIDs := make([]int, 0, len(permIDSet))
	for pid := range permIDSet {
		permIDs = append(permIDs, pid)
	}

	// Find matching combinations: general (sub_block_store_id IS NULL) or store-specific
	var combinations []permission.PermissionsCombination
	q := tx.Where("permissions_id IN ?", permIDs).
		Where("permissions_sub_block_warehouse_id IS NULL")
	if hasSubBlock {
		q = q.Where("permissions_sub_block_store_id IS NULL OR permissions_sub_block_store_id = ?",
			storeSubBlock.IDPermissionsSubBlock)
	} else {
		q = q.Where("permissions_sub_block_store_id IS NULL")
	}
	if err := q.Find(&combinations).Error; err != nil {
		return err
	}

	now := time.Now()
	for _, combo := range combinations {
		var existing employees.EmployeePermission
		err := tx.Where(
			"employee_login_id = ? AND permissions_combination_id = ?",
			employeeLoginID, combo.IDPermissionsCombination,
		).First(&existing).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			ep := employees.EmployeePermission{
				EmployeeLoginID:          employeeLoginID,
				PermissionsCombinationID: combo.IDPermissionsCombination,
				GrantedBy:                grantedBy,
				GrantedAt:                &now,
				IsActive:                 true,
			}
			if err := tx.Create(&ep).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !existing.IsActive {
			if err := tx.Model(&existing).Updates(map[string]interface{}{
				"is_active":  true,
				"granted_by": grantedBy,
				"granted_at": now,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

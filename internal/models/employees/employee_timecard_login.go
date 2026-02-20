package employees

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// EmployeeTimecardLogin ⇄ employee_timecard_login
type EmployeeTimecardLogin struct {
	IDEmployeeTimecardLogin int    `gorm:"column:id_employee_timecard_login;primaryKey;autoIncrement" json:"id_employee_timecard_login"`
	EmployeeID              *int   `gorm:"column:employee_id"                                        json:"employee_id,omitempty"`
	Username                string `gorm:"column:username;type:varchar(255);not null;uniqueIndex:uniq_timecard_username" json:"username"`
	PasswordHash            string `gorm:"column:password_hash;type:varchar(255);not null"           json:"-"`
	FirstName               string `gorm:"column:first_name;type:varchar(50);not null"               json:"first_name"`
	LastName                string `gorm:"column:last_name;type:varchar(50);not null"                json:"last_name"`
	Active                  bool   `gorm:"column:active;not null;default:true"                       json:"active"`
}

func (EmployeeTimecardLogin) TableName() string { return "employee_timecard_login" }

// SetPassword — hash and store password (bcrypt.DefaultCost)
func (e *EmployeeTimecardLogin) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	e.PasswordHash = string(hash)
	return nil
}

// CheckPassword — verify password against stored hash
func (e *EmployeeTimecardLogin) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(e.PasswordHash), []byte(password)) == nil
}

// ToMap — как Python to_dict()
func (e *EmployeeTimecardLogin) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_employee_timecard_login": e.IDEmployeeTimecardLogin,
		"employee_id":                e.EmployeeID,
		"username":                   e.Username,
		"first_name":                 e.FirstName,
		"last_name":                  e.LastName,
		"active":                     e.Active,
	}
}

func (e *EmployeeTimecardLogin) String() string {
	return fmt.Sprintf("<EmployeeTimecardLogin %s>", e.Username)
}

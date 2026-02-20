// internal/models/auth/employee_login_model.go
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// EmployeeLogin ⇄ employee_login
type EmployeeLogin struct {
	IDEmployeeLogin int     `gorm:"column:id_employee_login;primaryKey;autoIncrement"           json:"id_employee_login"`
	Username        string  `gorm:"column:employee_login;type:varchar(255);not null"            json:"employee_login"`
	PasswordHash    string  `gorm:"column:password_hash;type:varchar(255);not null"             json:"-"`
	ExpressLogin    *string `gorm:"column:express_login;type:varchar(255);uniqueIndex"          json:"-"`
	FailedAttempts  int     `gorm:"column:failed_attempts;not null;default:0"                   json:"failed_attempts"`
	Active          bool    `gorm:"column:active;not null;default:true"                         json:"active"`
}

func (EmployeeLogin) TableName() string { return "employee_login" }

// NewEmployeeLogin — аналог Python __init__(username, password, active=True)
func NewEmployeeLogin(username, password string, active bool) (*EmployeeLogin, error) {
	el := &EmployeeLogin{
		Username: username,
		Active:   active,
	}
	if err := el.SetPassword(password); err != nil {
		return nil, err
	}
	return el, nil
}

// SetPassword — bcrypt хэш пароля
func (e *EmployeeLogin) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	e.PasswordHash = string(hash)
	return nil
}

// CheckPassword — проверка пароля
func (e *EmployeeLogin) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(e.PasswordHash), []byte(password)) == nil
}

// SetExpressLogin — сохранить PIN как bcrypt-хэш
func (e *EmployeeLogin) SetExpressLogin(pin string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s := string(hash)
	e.ExpressLogin = &s
	return nil
}

// CheckExpressLogin — проверка PIN
func (e *EmployeeLogin) CheckExpressLogin(pin string) bool {
	if e.ExpressLogin == nil || *e.ExpressLogin == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(*e.ExpressLogin), []byte(pin)) == nil
}

func (e *EmployeeLogin) String() string {
	return fmt.Sprintf("<EmployeeLogin %s>", e.Username)
}

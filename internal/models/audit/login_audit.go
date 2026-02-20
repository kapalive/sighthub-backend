// internal/models/audit/login_audit.go
package audit

import (
	"fmt"
	"time"
)

// LoginAudit ⇄ login_audit
type LoginAudit struct {
	LogID        int       `gorm:"column:log_id;primaryKey"                                             json:"log_id"`
	LogTime      time.Time `gorm:"column:log_time;type:timestamptz;default:CURRENT_TIMESTAMP"           json:"log_time"`
	UserName     string    `gorm:"column:user_name;type:varchar(255);not null"                          json:"user_name"`
	LocationName *string   `gorm:"column:location_name;type:varchar(255)"                               json:"location_name,omitempty"`
	LoginMethod  *string   `gorm:"column:login_method;type:varchar(50)"                                 json:"login_method,omitempty"`
	IPAddress    *string   `gorm:"column:ip_address;type:varchar(50)"                                   json:"ip_address,omitempty"`
	BrowserType  *string   `gorm:"column:browser_type;type:varchar(50)"                                 json:"browser_type,omitempty"`
	LoginStatus  bool      `gorm:"column:login_status;not null"                                         json:"login_status"`
	PWStatus     *bool     `gorm:"column:pw_status"                                                      json:"pw_status,omitempty"`
	ActiveStatus *bool     `gorm:"column:active_status"                                                 json:"active_status,omitempty"`
	EmployeeID   *int      `gorm:"column:employee_id"                                                   json:"employee_id,omitempty"`
}

func (LoginAudit) TableName() string { return "login_audit" }

func (l *LoginAudit) ToMap() map[string]interface{} {
	var ts string
	if !l.LogTime.IsZero() {
		ts = l.LogTime.Format(time.RFC3339)
	}
	return map[string]interface{}{
		"log_id":        l.LogID,
		"log_time":      ts,
		"user_name":     l.UserName,
		"location_name": l.LocationName,
		"login_method":  l.LoginMethod,
		"ip_address":    l.IPAddress,
		"browser_type":  l.BrowserType,
		"login_status":  l.LoginStatus,
		"pw_status":     l.PWStatus,
		"active_status": l.ActiveStatus,
		"employee_id":   l.EmployeeID,
	}
}

func (l *LoginAudit) String() string {
	return fmt.Sprintf("<LoginAudit %d %s>", l.LogID, l.UserName)
}

// internal/models/permission/roles.go
package permission

// Role ⇄ table: roles
type Role struct {
	RoleID      int     `gorm:"column:role_id;primaryKey;autoIncrement"              json:"role_id"`
	RoleName    string  `gorm:"column:role_name;type:varchar(255);not null;uniqueIndex" json:"role_name"`
	Description *string `gorm:"column:description;type:text"                          json:"description,omitempty"`
	Key         string  `gorm:"column:key;type:varchar(127);not null;uniqueIndex"     json:"key"`
}

func (Role) TableName() string { return "roles" }

func (r *Role) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"role_id":     r.RoleID,
		"role_name":   r.RoleName,
		"description": r.Description,
		"key":         r.Key,
	}
}

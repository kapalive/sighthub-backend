// internal/models/permission/role_permissions.go
package permission

// RolePermission ⇄ table: role_permissions
type RolePermission struct {
	RolePermissionID int `gorm:"column:role_permission_id;primaryKey;autoIncrement"           json:"role_permission_id"`
	RoleID           int `gorm:"column:role_id;not null;uniqueIndex:uix_role_permission"      json:"role_id"`
	PermissionsID    int `gorm:"column:permissions_id;not null;uniqueIndex:uix_role_permission" json:"permissions_id"`
}

func (RolePermission) TableName() string { return "role_permissions" }

func (r *RolePermission) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"role_permission_id": r.RolePermissionID,
		"role_id":            r.RoleID,
		"permissions_id":     r.PermissionsID,
	}
}

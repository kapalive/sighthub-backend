// internal/models/permission/permissions.go
package permission

// Permissions ⇄ table: permissions
type Permissions struct {
	PermissionsID                  int     `gorm:"column:permissions_id;primaryKey"                              json:"permissions_id"`
	PermissionName                 string  `gorm:"column:permission_name;type:varchar(255);not null;uniqueIndex" json:"permission_name"`
	Description                    *string `gorm:"column:description;type:text"                                  json:"description,omitempty"`
	PermissionsBlockID             int     `gorm:"column:permissions_block_id;not null"                          json:"permissions_block_id"`
	PermissionsSubBlockStoreID     *int    `gorm:"column:permissions_sub_block_store_id"                         json:"permissions_sub_block_store_id,omitempty"`
	PermissionsSubBlockWarehouseID *int    `gorm:"column:permissions_sub_block_warehouse_id"                     json:"permissions_sub_block_warehouse_id,omitempty"`
}

func (Permissions) TableName() string { return "permissions" }

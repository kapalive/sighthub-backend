// internal/models/permission/permissions_combination.go
package permission

// PermissionsCombination ⇄ table: permissions_combination
type PermissionsCombination struct {
	IDPermissionsCombination       int  `gorm:"column:id_permissions_combination;primaryKey;autoIncrement" json:"id_permissions_combination"`
	PermissionsBlockID             *int `gorm:"column:permissions_block_id"                                json:"permissions_block_id,omitempty"`
	PermissionsSubBlockStoreID     *int `gorm:"column:permissions_sub_block_store_id"                      json:"permissions_sub_block_store_id,omitempty"`
	PermissionsSubBlockWarehouseID *int `gorm:"column:permissions_sub_block_warehouse_id"                  json:"permissions_sub_block_warehouse_id,omitempty"`
	PermissionsID                  int  `gorm:"column:permissions_id;not null"                             json:"permissions_id"`
}

func (PermissionsCombination) TableName() string { return "permissions_combination" }

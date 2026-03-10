// internal/models/permission/permissions_sub_block_warehouse.go
package permission

// PermissionsSubBlockWarehouse ⇄ table: permissions_sub_block_warehouse
type PermissionsSubBlockWarehouse struct {
	IDPermissionsSubBlock int     `gorm:"column:id_permissions_sub_block;primaryKey;autoIncrement"      json:"id_permissions_sub_block"`
	SubBlockName          string  `gorm:"column:sub_block_name;type:varchar(255);not null;uniqueIndex"  json:"sub_block_name"`
	Description           *string `gorm:"column:description;type:text"                                  json:"description,omitempty"`
	WarehouseID           *int    `gorm:"column:warehouse_id"                                           json:"warehouse_id,omitempty"`
}

func (PermissionsSubBlockWarehouse) TableName() string { return "permissions_sub_block_warehouse" }

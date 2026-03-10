// internal/models/permission/permissions_sub_block_store.go
package permission

// PermissionsSubBlockStore ⇄ table: permissions_sub_block_store
type PermissionsSubBlockStore struct {
	IDPermissionsSubBlock int     `gorm:"column:id_permissions_sub_block;primaryKey;autoIncrement"      json:"id_permissions_sub_block"`
	SubBlockName          string  `gorm:"column:sub_block_name;type:varchar(255);not null;uniqueIndex"  json:"sub_block_name"`
	Description           *string `gorm:"column:description;type:text"                                  json:"description,omitempty"`
	StoreID               *int    `gorm:"column:store_id"                                               json:"store_id,omitempty"`
}

func (PermissionsSubBlockStore) TableName() string { return "permissions_sub_block_store" }

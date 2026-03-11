package permission

type PermissionsBlock struct {
	IDPermissionsBlock int     `gorm:"column:id_permissions_block;primaryKey" json:"id_permissions_block"`
	BlockName          string  `gorm:"column:block_name;size:255;not null;uniqueIndex" json:"block_name"`
	Description        *string `gorm:"column:description;type:text"                   json:"description,omitempty"`
}

func (PermissionsBlock) TableName() string { return "permissions_block" }

package permission

type TableAccess struct {
	TableAccessID int    `gorm:"column:table_access_id;primaryKey;autoIncrement" json:"table_access_id"`
	RoleID        int    `gorm:"column:role_id;not null;uniqueIndex:uix_table_access" json:"role_id"`
	Table         string `gorm:"column:table_name;size:255;not null;uniqueIndex:uix_table_access" json:"table_name"`
	PermissionsID int    `gorm:"column:permissions_id;not null;uniqueIndex:uix_table_access" json:"permissions_id"`
}

func (TableAccess) TableName() string { return "table_access" }

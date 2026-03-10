// internal/models/vendors/vendor_location_account.go
package vendors

import "time"

// VendorLocationAccount ⇄ vendor_location_account
type VendorLocationAccount struct {
	IDVendorLocationAccount int64      `gorm:"column:id_vendor_location_account;primaryKey;autoIncrement" json:"id_vendor_location_account"`
	VendorID                int        `gorm:"column:vendor_id;not null"                                  json:"vendor_id"`
	LocationID              int64      `gorm:"column:location_id;not null"                                json:"location_id"`
	AccountNumber           string     `gorm:"column:account_number;type:varchar(64);not null"            json:"account_number"`
	QbVendorRef             *string    `gorm:"column:qb_vendor_ref;type:varchar(64)"                      json:"qb_vendor_ref,omitempty"`
	Note                    *string    `gorm:"column:note;type:varchar(255)"                              json:"note,omitempty"`
	IsActive                bool       `gorm:"column:is_active;not null;default:true"                     json:"is_active"`
	Status                  *string    `gorm:"column:status;type:varchar(20)"                             json:"status,omitempty"`
	CreatedAt               *time.Time `gorm:"column:created_at;type:timestamptz;default:now()"           json:"created_at,omitempty"`
	UpdatedAt               *time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"           json:"updated_at,omitempty"`
}

func (VendorLocationAccount) TableName() string { return "vendor_location_account" }

func (v *VendorLocationAccount) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_vendor_location_account": v.IDVendorLocationAccount,
		"vendor_id":                  v.VendorID,
		"location_id":                v.LocationID,
		"account_number":             v.AccountNumber,
		"qb_vendor_ref":              v.QbVendorRef,
		"note":                       v.Note,
		"is_active":                  v.IsActive,
		"status":                     v.Status,
	}
	if v.CreatedAt != nil {
		m["created_at"] = v.CreatedAt.Format(time.RFC3339)
	} else {
		m["created_at"] = nil
	}
	if v.UpdatedAt != nil {
		m["updated_at"] = v.UpdatedAt.Format(time.RFC3339)
	} else {
		m["updated_at"] = nil
	}
	return m
}

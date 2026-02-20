// internal/models/vendors/agreement.go
package vendors

import (
	"fmt"
	"time"
)

type Agreement struct {
	IDAgreement   int        `gorm:"column:id_agreement;primaryKey"      json:"id_agreement"`
	LinkToFile    *string    `gorm:"column:link_to_file;type:text"        json:"link_to_file,omitempty"`
	Title         *string    `gorm:"column:title;type:varchar(255)"       json:"title,omitempty"`
	DateAgreement *time.Time `gorm:"column:date_agreement;type:date"      json:"-"`
	DateEnd       *time.Time `gorm:"column:date_end;type:date"            json:"-"`
	VendorID      int        `gorm:"column:vendor_id;not null;index"      json:"vendor_id"`

	Vendor *Vendor `gorm:"foreignKey:VendorID;references:IDVendor" json:"-"`
}

func (Agreement) TableName() string { return "agreement" }

func (a *Agreement) ToMap() map[string]interface{} {
	var da any
	if a.DateAgreement != nil {
		da = a.DateAgreement.Format("2006-01-02")
	}
	var de any
	if a.DateEnd != nil {
		de = a.DateEnd.Format("2006-01-02")
	}

	m := map[string]interface{}{
		"id_agreement":   a.IDAgreement,
		"link_to_file":   a.LinkToFile,
		"title":          a.Title,
		"date_agreement": da,
		"date_end":       de,
	}

	if a.Vendor != nil {
		m["vendor"] = a.Vendor.ToMap()
	} else {
		m["vendor"] = nil
	}
	return m
}

func (a *Agreement) String() string {
	return fmt.Sprintf("<Agreement %d | Start: %v | End: %v>", a.IDAgreement, a.DateAgreement, a.DateEnd)
}

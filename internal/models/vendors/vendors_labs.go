// internal/models/vendors/vendors_labs.go
package vendors

import "fmt"

type VendorLabs struct {
	IDVendorLabs int `gorm:"column:id_vendor_labs;primaryKey" json:"id_vendor_labs"`

	VendorID int `gorm:"column:vendor_id;not null;index;uniqueIndex:uq_vendor_labs_vendor_id_lab_id" json:"vendor_id"`
	LabID    int `gorm:"column:lab_id;not null;index;uniqueIndex:uq_vendor_labs_vendor_id_lab_id"    json:"lab_id"`

	Vendor *Vendor `gorm:"foreignKey:VendorID;references:IDVendor" json:"-"`
	Lab    *Lab    `gorm:"foreignKey:LabID;references:IDLab"       json:"-"`
}

func (VendorLabs) TableName() string { return "vendor_labs" }

func (vl *VendorLabs) String() string {
	return fmt.Sprintf("<VendorLabs id_vendor_labs: %d, vendor_id: %d, lab_id: %d>", vl.IDVendorLabs, vl.VendorID, vl.LabID)
}

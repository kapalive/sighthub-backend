// internal/models/vendors/rep_contact_vendors.go
package vendors

import (
	"fmt"
	"sighthub-backend/internal/models/general"
)

type RepContactVendor struct {
	IDRepContactVendor int `gorm:"column:id_rep_contact_vendor;primaryKey"                         json:"id_rep_contact_vendor"`
	VendorID           int `gorm:"column:vendor_id;not null;index;constraint:OnDelete:CASCADE;"    json:"vendor_id"`

	Name          *string `gorm:"column:name;type:varchar(100)"            json:"name,omitempty"`
	Title         *string `gorm:"column:title;type:varchar(100)"           json:"title,omitempty"`
	Phone         *string `gorm:"column:phone;type:varchar(20)"            json:"phone,omitempty"`
	AddressLine2  *string `gorm:"column:address_line_2;type:varchar(255)"  json:"address_line_2,omitempty"`
	City          *string `gorm:"column:city;type:varchar(100)"            json:"city,omitempty"`
	Country       *string `gorm:"column:country;type:varchar(100)"         json:"-"`
	Email         *string `gorm:"column:email;type:varchar(100)"           json:"email,omitempty"`
	Fax           *string `gorm:"column:fax;type:varchar(20)"              json:"fax,omitempty"`
	State         *string `gorm:"column:state;type:varchar(100)"           json:"-"`
	StreetAddress *string `gorm:"column:street_address;type:varchar(255)"  json:"street_address,omitempty"`
	Zip           *string `gorm:"column:zip;type:varchar(20)"              json:"zip,omitempty"`

	CountryID *int `gorm:"column:country_id"   json:"country_id,omitempty"`
	StateID   *int `gorm:"column:state_id"     json:"state_id,omitempty"`

	Vendor     *Vendor                  `gorm:"foreignKey:VendorID;references:IDVendor"             json:"-"`
	CountryRef *general.Country         `gorm:"foreignKey:CountryID;references:IDCountry"           json:"-"`
	StateRef   *general.SalesTaxByState `gorm:"foreignKey:StateID;references:IDSalesTax"            json:"-"`
}

func (RepContactVendor) TableName() string { return "rep_contact_vendor" }

func (r *RepContactVendor) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_rep_contact_vendor": r.IDRepContactVendor,
		"vendor_id":             r.VendorID,
		"name":                  r.Name,
		"title":                 r.Title,
		"phone":                 r.Phone,
		"address_line_2":        r.AddressLine2,
		"city":                  r.City,
		"email":                 r.Email,
		"fax":                   r.Fax,
		"street_address":        r.StreetAddress,
		"zip":                   r.Zip,
	}
	if r.CountryID != nil {
		m["country_id"] = *r.CountryID
	} else {
		m["country_id"] = nil
	}
	if r.StateID != nil {
		m["state_id"] = *r.StateID
	} else {
		m["state_id"] = nil
	}
	return m
}

func (r *RepContactVendor) String() string {
	return fmt.Sprintf("<RepContactVendor %d vendor=%d>", r.IDRepContactVendor, r.VendorID)
}

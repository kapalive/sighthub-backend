// internal/models/vendors/vendors.go
package vendors

import "fmt"

type Vendor struct {
	IDVendor        int     `gorm:"column:id_vendor;primaryKey"              json:"id_vendor"`
	VendorName      string  `gorm:"column:vendor_name;type:varchar(100);not null" json:"vendor_name"`
	ShortName       *string `gorm:"column:short_name;type:varchar(20)"       json:"short_name,omitempty"`
	Phone           *string `gorm:"column:phone;type:varchar(20)"            json:"phone,omitempty"`
	ExtPhone        *string `gorm:"column:ext_phone;type:varchar(20)"        json:"ext_phone,omitempty"`
	StreetAddress   *string `gorm:"column:street_address;type:varchar(100)"  json:"street_address,omitempty"`
	AddressLine2    *string `gorm:"column:address_line_2;type:varchar(100)"  json:"address_line_2,omitempty"`
	City            *string `gorm:"column:city;type:varchar(100)"            json:"city,omitempty"`
	State           *string `gorm:"column:state;type:varchar(255)"           json:"state,omitempty"`
	ZipCode         *string `gorm:"column:zip_code;type:varchar(7)"          json:"zip_code,omitempty"`
	Country         *string `gorm:"column:country;type:varchar(255)"         json:"country,omitempty"`
	Website         *string `gorm:"column:website;type:varchar(100)"         json:"website,omitempty"`
	Fax             *string `gorm:"column:fax;type:varchar(20)"              json:"fax,omitempty"`
	Email           *string `gorm:"column:email;type:varchar(100)"           json:"email,omitempty"`
	RegionalManager *string `gorm:"column:regional_manager;type:varchar(100)" json:"regional_manager,omitempty"`
	RegionalMNo     *string `gorm:"column:regional_m_no;type:varchar(20)"    json:"regional_m_no,omitempty"`

	Frames        bool `gorm:"column:frames;not null;default:false"         json:"frames"`
	Lenses        bool `gorm:"column:lenses;not null;default:false"         json:"lenses"`
	ContactLenses bool `gorm:"column:contact_lenses;not null;default:false" json:"contact_lenses"`

	CountryID *int `gorm:"column:country_id"            json:"country_id,omitempty"`
	StateID   *int `gorm:"column:state_id"              json:"state_id,omitempty"`
}

func (Vendor) TableName() string { return "vendor" }

func (v *Vendor) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_vendor":      v.IDVendor,
		"vendor_name":    v.VendorName,
		"short_name":     v.ShortName,
		"phone":          v.Phone,
		"street_address": v.StreetAddress,
		"address_line_2": v.AddressLine2,
		"city":           v.City,
		"state_id":       v.StateID,
		"zip_code":       v.ZipCode,
		"country_id":     v.CountryID,
		"website":        v.Website,
		"fax":            v.Fax,
		"email":          v.Email,
		"frames":         v.Frames,
		"lenses":         v.Lenses,
		"contact_lenses": v.ContactLenses,
	}
}

func (v *Vendor) String() string {
	city := ""
	if v.City != nil {
		city = *v.City
	}
	state := ""
	if v.State != nil {
		state = *v.State
	}
	return fmt.Sprintf("<Vendor %s | %s, %s>", v.VendorName, city, state)
}

func (v *Vendor) Name() string {
	return v.VendorName
}

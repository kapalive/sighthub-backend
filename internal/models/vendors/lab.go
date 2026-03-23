// internal/models/vendors/lab.go
package vendors

import (
	"fmt"
	"sighthub-backend/internal/models/general"
)

type Lab struct {
	IDLab         int     `gorm:"column:id_lab;primaryKey"                         json:"id_lab"`
	TitleLab      string  `gorm:"column:title_lab;type:varchar(100);not null"      json:"title_lab"`
	ShortName     *string `gorm:"column:short_name;type:varchar(3)"                json:"short_name,omitempty"`
	IsInternal    bool    `gorm:"column:is_internal;not null;default:false"        json:"is_internal"`
	Phone         *string `gorm:"column:phone;type:varchar(13)"                    json:"phone,omitempty"`
	Email         *string `gorm:"column:email;type:varchar(150)"                   json:"email,omitempty"`
	StreetAddress *string `gorm:"column:street_address;type:varchar(100)"         json:"street_address,omitempty"`
	AddressLine2  *string `gorm:"column:address_line_2;type:varchar(100)"          json:"address_line_2,omitempty"`
	City          *string `gorm:"column:city;type:varchar(100)"                    json:"city,omitempty"`
	ZipCode       *string `gorm:"column:zip_code;type:varchar(7)"                  json:"zip_code,omitempty"`

	StateID       *int    `gorm:"column:state_id"   json:"state_id,omitempty"`
	CountryID     *int    `gorm:"column:country_id" json:"country_id,omitempty"`
	VendorID      *int    `gorm:"column:vendor_id"  json:"vendor_id,omitempty"`
	BrandLensID   *int    `gorm:"column:brand_lens_id" json:"brand_lens_id,omitempty"`
	Source        *string `gorm:"column:source;type:varchar(20)" json:"source,omitempty"`

	State   *general.SalesTaxByState `gorm:"foreignKey:StateID;references:IDSalesTax" json:"-"`
	Country *general.Country         `gorm:"foreignKey:CountryID;references:IDCountry" json:"-"`
}

func (Lab) TableName() string { return "lab" }

func (l *Lab) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_lab":         l.IDLab,
		"title_lab":      l.TitleLab,
		"short_name":     l.ShortName,
		"is_internal":    l.IsInternal,
		"phone":          l.Phone,
		"email":          l.Email,
		"street_address": l.StreetAddress,
		"address_line_2": l.AddressLine2,
		"city":           l.City,
		"zip_code":       l.ZipCode,
		"vendor_id":      l.VendorID,
		"brand_lens_id":  l.BrandLensID,
		"source":         l.Source,
	}

	if l.StateID != nil {
		m["state_id"] = *l.StateID
	} else {
		m["state_id"] = nil
	}
	if l.CountryID != nil {
		m["country_id"] = *l.CountryID
	} else {
		m["country_id"] = nil
	}

	if l.State != nil {
		m["state"] = l.State.StateName
	} else {
		m["state"] = nil
	}
	if l.Country != nil {
		m["country"] = l.Country.Country
	} else {
		m["country"] = nil
	}

	return m
}

func (l *Lab) String() string {
	var sn string
	if l.ShortName != nil {
		sn = *l.ShortName
	}
	return fmt.Sprintf("<Lab %s | %s>", l.TitleLab, sn)
}

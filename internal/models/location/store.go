// internal/models/location/store.go
package location

// Store ⇄ table: store
type Store struct {
	IDStore       int     `gorm:"column:id_store;primaryKey;autoIncrement"          json:"id_store"`
	FullName      *string `gorm:"column:full_name;type:varchar(100)"                json:"full_name,omitempty"`
	ShortName     *string `gorm:"column:short_name;type:varchar(2)"                 json:"short_name,omitempty"`
	StreetAddress *string `gorm:"column:street_address;type:varchar(100)"           json:"street_address,omitempty"`
	AddressLine2  *string `gorm:"column:address_line_2;type:varchar(100)"           json:"address_line_2,omitempty"`
	City          *string `gorm:"column:city;type:varchar(100)"                     json:"city,omitempty"`
	State         *string `gorm:"column:state;type:varchar(2)"                      json:"state,omitempty"`
	PostalCode    *string `gorm:"column:postal_code;type:varchar(20)"               json:"postal_code,omitempty"`
	Country       *string `gorm:"column:country;type:varchar(50)"                   json:"country,omitempty"`
	Phone         *string `gorm:"column:phone;type:varchar(20)"                     json:"phone,omitempty"`
	TimeZone      *string `gorm:"column:time_zone;type:varchar(200)"                json:"time_zone,omitempty"`
	Fax           *string `gorm:"column:fax;type:varchar(20)"                       json:"fax,omitempty"`
	Email         *string `gorm:"column:email;type:varchar(100)"                    json:"email,omitempty"`
	BusinessName  *string `gorm:"column:business_name;type:varchar(255)"            json:"business_name,omitempty"`
	NPI           *string `gorm:"column:npi;type:varchar(20)"                       json:"npi,omitempty"`
	TaxN          *string `gorm:"column:tax_n;type:varchar(20)"                     json:"tax_n,omitempty"`
	HPSA          *string `gorm:"column:hpsa;type:varchar(100)"                     json:"hpsa,omitempty"`
	Logo          *string `gorm:"column:logo;type:text"                             json:"logo,omitempty"`
	LicenseKey    *string `gorm:"column:license_key;type:text"                      json:"license_key,omitempty"`
	Hash          string  `gorm:"column:hash;type:varchar(32);uniqueIndex;not null" json:"hash"`
}

func (Store) TableName() string { return "store" }

func (s *Store) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_store":       s.IDStore,
		"full_name":      s.FullName,
		"short_name":     s.ShortName,
		"street_address": s.StreetAddress,
		"address_line_2": s.AddressLine2,
		"city":           s.City,
		"state":          s.State,
		"postal_code":    s.PostalCode,
		"country":        s.Country,
		"phone":          s.Phone,
		"time_zone":      s.TimeZone,
		"fax":            s.Fax,
		"email":          s.Email,
		"business_name":  s.BusinessName,
		"npi":            s.NPI,
		"tax_n":          s.TaxN,
		"hpsa":           s.HPSA,
		"logo":           s.Logo,
		"license_key":    s.LicenseKey,
		"hash":           s.Hash,
	}
}

// internal/models/location/location.go
package location

import "strconv"

// salesTaxRef — минимальная inline-структура для preload в Location.
// Полноценная модель SalesTaxByState — в internal/models/general.
type salesTaxRef struct {
	IDSalesTax      int      `gorm:"column:id_sales_tax;primaryKey"`
	SalesTaxPercent *float64 `gorm:"column:sales_tax_percent"`
}

func (salesTaxRef) TableName() string { return "sales_tax_by_state" }

// Location ⇄ table: location
type Location struct {
	IDLocation int `gorm:"column:id_location;primaryKey" json:"id_location"`

	FullName      string  `gorm:"column:full_name;type:varchar(100);not null" json:"full_name"`
	ShortName     *string `gorm:"column:short_name;type:varchar(2)"           json:"short_name,omitempty"`
	StreetAddress *string `gorm:"column:street_address;type:varchar(100)"     json:"street_address,omitempty"`
	AddressLine2  *string `gorm:"column:address_line_2;type:varchar(100)"     json:"address_line_2,omitempty"`
	City          *string `gorm:"column:city;type:varchar(100)"               json:"city,omitempty"`
	State         *string `gorm:"column:state;type:varchar(2)"                json:"state,omitempty"`
	PostalCode    *string `gorm:"column:postal_code;type:varchar(20)"         json:"postal_code,omitempty"`
	Country       *string `gorm:"column:country;type:varchar(50)"             json:"country,omitempty"`
	Phone         *string `gorm:"column:phone;type:varchar(20)"               json:"phone,omitempty"`
	Website       *string `gorm:"column:website;type:varchar(100)"            json:"website,omitempty"`
	Fax           *string `gorm:"column:fax;type:varchar(20)"                 json:"fax,omitempty"`
	Email         *string `gorm:"column:email;type:varchar(100)"              json:"email,omitempty"`
	TimeZone      *string `gorm:"column:time_zone;type:varchar(200)"          json:"time_zone,omitempty"`
	WorkingHours  *string `gorm:"column:working_hours;type:varchar(200)"      json:"working_hours,omitempty"`

	SalesTaxID  *int `gorm:"column:sales_tax_id"      json:"sales_tax_id,omitempty"`
	StoreID     int  `gorm:"column:store_id;not null" json:"store_id"`
	WarehouseID *int `gorm:"column:warehouse_id"      json:"warehouse_id,omitempty"`
	WorkShiftID *int `gorm:"column:work_shift_id"     json:"work_shift_id,omitempty"`

	CanReceiveItems *bool `gorm:"column:can_receive_items;default:true"  json:"can_receive_items,omitempty"`
	StoreActive     *bool `gorm:"column:store_active;default:false"      json:"store_active,omitempty"`
	Showcase        *bool `gorm:"column:showcase;default:false"          json:"showcase,omitempty"`

	LogoPath *string `gorm:"column:logo_path;type:varchar(255)" json:"logo_path,omitempty"`

	// Relations — используются только при Preload
	SalesTax  *salesTaxRef `gorm:"foreignKey:SalesTaxID;references:IDSalesTax"      json:"-"`
	Store     *Store       `gorm:"foreignKey:StoreID;references:IDStore"            json:"-"`
	Warehouse *Warehouse   `gorm:"foreignKey:WarehouseID;references:IDWarehouse"    json:"-"`
}

func (Location) TableName() string { return "location" }

// ToMap — аналог Python to_dict()
func (l *Location) ToMap() map[string]interface{} {
	var storeShort, storeFull interface{}
	if l.Store != nil {
		storeShort = l.Store.ShortName
		storeFull = l.Store.FullName
	}

	var whShort, whFull interface{}
	if l.Warehouse != nil {
		whShort = l.Warehouse.ShortName
		whFull = l.Warehouse.FullName
	}

	var taxPercent interface{}
	if l.SalesTax != nil && l.SalesTax.SalesTaxPercent != nil {
		taxPercent = strconv.FormatFloat(*l.SalesTax.SalesTaxPercent, 'f', -1, 64)
	}

	return map[string]interface{}{
		"id_location":    l.IDLocation,
		"full_name":      l.FullName,
		"short_name":     l.ShortName,
		"street_address": l.StreetAddress,
		"address_line_2": l.AddressLine2,
		"city":           l.City,
		"state":          l.State,
		"postal_code":    l.PostalCode,
		"country":        l.Country,
		"phone":          l.Phone,
		"website":        l.Website,
		"fax":            l.Fax,
		"email":          l.Email,
		"time_zone":      l.TimeZone,
		"working_hours":  l.WorkingHours,

		"can_receive_items": l.CanReceiveItems,
		"store_active":      l.StoreActive,
		"showcase":          l.Showcase,

		"store_id":         l.StoreID,
		"store_short_name": storeShort,
		"store_full_name":  storeFull,

		"warehouse_id":         l.WarehouseID,
		"warehouse_short_name": whShort,
		"warehouse_full_name":  whFull,

		"sales_tax_percent": taxPercent,
		"logo_path":         l.LogoPath,
	}
}

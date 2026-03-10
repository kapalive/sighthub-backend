// internal/models/prescriptions/contact_lens_prescription.go
package prescriptions

import "time"

// ContactLensPrescription ↔ table: contact_lens_prescription
type ContactLensPrescription struct {
	IDContactLensPrescription int64      `gorm:"column:id_contact_lens_prescription;primaryKey;autoIncrement" json:"id_contact_lens_prescription"`
	PrescriptionID            int64      `gorm:"column:prescription_id;not null"                              json:"prescription_id"`
	OdContLens                *string    `gorm:"column:od_cont_lens;type:varchar(200)"                        json:"od_cont_lens,omitempty"`
	OsContLens                *string    `gorm:"column:os_cont_lens;type:varchar(200)"                        json:"os_cont_lens,omitempty"`
	OdBc                      *string    `gorm:"column:od_bc;type:varchar(20)"                                json:"od_bc,omitempty"`
	OsBc                      *string    `gorm:"column:os_bc;type:varchar(20)"                                json:"os_bc,omitempty"`
	OdDia                     *float64   `gorm:"column:od_dia;type:numeric(5,2)"                              json:"od_dia,omitempty"`
	OsDia                     *float64   `gorm:"column:os_dia;type:numeric(5,2)"                              json:"os_dia,omitempty"`
	OdPwr                     *string    `gorm:"column:od_pwr;type:varchar(20)"                               json:"od_pwr,omitempty"`
	OsPwr                     *string    `gorm:"column:os_pwr;type:varchar(20)"                               json:"os_pwr,omitempty"`
	OdCyl                     *string    `gorm:"column:od_cyl;type:varchar(20)"                               json:"od_cyl,omitempty"`
	OsCyl                     *string    `gorm:"column:os_cyl;type:varchar(20)"                               json:"os_cyl,omitempty"`
	OdAxis                    *string    `gorm:"column:od_axis;type:varchar(20)"                              json:"od_axis,omitempty"`
	OsAxis                    *string    `gorm:"column:os_axis;type:varchar(20)"                              json:"os_axis,omitempty"`
	OdAdd                     *string    `gorm:"column:od_add;type:varchar(20)"                               json:"od_add,omitempty"`
	OsAdd                     *string    `gorm:"column:os_add;type:varchar(20)"                               json:"os_add,omitempty"`
	OdColor                   *string    `gorm:"column:od_color;type:varchar(100)"                            json:"od_color,omitempty"`
	OsColor                   *string    `gorm:"column:os_color;type:varchar(100)"                            json:"os_color,omitempty"`
	OdType                    *string    `gorm:"column:od_type;type:varchar(20)"                              json:"od_type,omitempty"` // "Patient Rx" | "Trial"
	OsType                    *string    `gorm:"column:os_type;type:varchar(20)"                              json:"os_type,omitempty"`
	ExpirationDate            *time.Time `gorm:"column:expiration_date;type:date"                             json:"expiration_date,omitempty"`
	OdHPrismDirection         *string    `gorm:"column:od_h_prism_direction;type:varchar(2)"                  json:"od_h_prism_direction,omitempty"` // BI | BO
	OsHPrismDirection         *string    `gorm:"column:os_h_prism_direction;type:varchar(2)"                  json:"os_h_prism_direction,omitempty"`
	OdVPrismDirection         *string    `gorm:"column:od_v_prism_direction;type:varchar(2)"                  json:"od_v_prism_direction,omitempty"` // BU | BD
	OsVPrismDirection         *string    `gorm:"column:os_v_prism_direction;type:varchar(2)"                  json:"os_v_prism_direction,omitempty"`
}

func (ContactLensPrescription) TableName() string { return "contact_lens_prescription" }

func (c *ContactLensPrescription) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_contact_lens_prescription": c.IDContactLensPrescription,
		"prescription_id":              c.PrescriptionID,
		"od_cont_lens":                 c.OdContLens,
		"os_cont_lens":                 c.OsContLens,
		"od_bc":                        c.OdBc,
		"os_bc":                        c.OsBc,
		"od_dia":                       c.OdDia,
		"os_dia":                       c.OsDia,
		"od_pwr":                       c.OdPwr,
		"os_pwr":                       c.OsPwr,
		"od_cyl":                       c.OdCyl,
		"os_cyl":                       c.OsCyl,
		"od_axis":                      c.OdAxis,
		"os_axis":                      c.OsAxis,
		"od_add":                       c.OdAdd,
		"os_add":                       c.OsAdd,
		"od_color":                     c.OdColor,
		"os_color":                     c.OsColor,
		"od_type":                      c.OdType,
		"os_type":                      c.OsType,
		"od_h_prism_direction":         c.OdHPrismDirection,
		"os_h_prism_direction":         c.OsHPrismDirection,
		"od_v_prism_direction":         c.OdVPrismDirection,
		"os_v_prism_direction":         c.OsVPrismDirection,
	}
	if c.ExpirationDate != nil {
		m["expiration_date"] = c.ExpirationDate.Format("2006-01-02")
	} else {
		m["expiration_date"] = nil
	}
	return m
}

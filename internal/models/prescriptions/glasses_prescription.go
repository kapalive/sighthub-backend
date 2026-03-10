// internal/models/prescriptions/glasses_prescription.go
package prescriptions

import "time"

// GlassesPrescription ↔ table: glasses_prescription
type GlassesPrescription struct {
	IDGlassesPrescription int64      `gorm:"column:id_glasses_prescription;primaryKey;autoIncrement" json:"id_glasses_prescription"`
	PrescriptionID        int64      `gorm:"column:prescription_id;not null"                         json:"prescription_id"`
	OdSph                 *string    `gorm:"column:od_sph;type:varchar(20)"                          json:"od_sph,omitempty"`
	OsSph                 *string    `gorm:"column:os_sph;type:varchar(20)"                          json:"os_sph,omitempty"`
	OdCyl                 *string    `gorm:"column:od_cyl;type:varchar(20)"                          json:"od_cyl,omitempty"`
	OsCyl                 *string    `gorm:"column:os_cyl;type:varchar(20)"                          json:"os_cyl,omitempty"`
	OdAxis                *string    `gorm:"column:od_axis;type:varchar(20)"                         json:"od_axis,omitempty"`
	OsAxis                *string    `gorm:"column:os_axis;type:varchar(20)"                         json:"os_axis,omitempty"`
	OdAdd                 *float64   `gorm:"column:od_add;type:numeric(5,2)"                         json:"od_add,omitempty"`
	OsAdd                 *float64   `gorm:"column:os_add;type:numeric(5,2)"                         json:"os_add,omitempty"`
	OdHPrism              *float64   `gorm:"column:od_h_prism;type:numeric(5,2)"                     json:"od_h_prism,omitempty"`
	OsHPrism              *float64   `gorm:"column:os_h_prism;type:numeric(5,2)"                     json:"os_h_prism,omitempty"`
	OdHPrismDirection     *string    `gorm:"column:od_h_prism_direction;type:varchar(2)"             json:"od_h_prism_direction,omitempty"` // BI | BO
	OsHPrismDirection     *string    `gorm:"column:os_h_prism_direction;type:varchar(2)"             json:"os_h_prism_direction,omitempty"`
	OdVPrism              *float64   `gorm:"column:od_v_prism;type:numeric(5,2)"                     json:"od_v_prism,omitempty"`
	OsVPrism              *float64   `gorm:"column:os_v_prism;type:numeric(5,2)"                     json:"os_v_prism,omitempty"`
	OdVPrismDirection     *string    `gorm:"column:od_v_prism_direction;type:varchar(2)"             json:"od_v_prism_direction,omitempty"` // BU | BD
	OsVPrismDirection     *string    `gorm:"column:os_v_prism_direction;type:varchar(2)"             json:"os_v_prism_direction,omitempty"`
	OdDpd                 *float64   `gorm:"column:od_dpd;type:numeric(5,2)"                         json:"od_dpd,omitempty"`
	OsDpd                 *float64   `gorm:"column:os_dpd;type:numeric(5,2)"                         json:"os_dpd,omitempty"`
	ExpirationDate        *time.Time `gorm:"column:expiration_date;type:date"                        json:"expiration_date,omitempty"`
}

func (GlassesPrescription) TableName() string { return "glasses_prescription" }

func (g *GlassesPrescription) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_glasses_prescription": g.IDGlassesPrescription,
		"prescription_id":         g.PrescriptionID,
		"od_sph":                  g.OdSph,
		"os_sph":                  g.OsSph,
		"od_cyl":                  g.OdCyl,
		"os_cyl":                  g.OsCyl,
		"od_axis":                 g.OdAxis,
		"os_axis":                 g.OsAxis,
		"od_add":                  g.OdAdd,
		"os_add":                  g.OsAdd,
		"od_h_prism":              g.OdHPrism,
		"os_h_prism":              g.OsHPrism,
		"od_h_prism_direction":    g.OdHPrismDirection,
		"os_h_prism_direction":    g.OsHPrismDirection,
		"od_v_prism":              g.OdVPrism,
		"os_v_prism":              g.OsVPrism,
		"od_v_prism_direction":    g.OdVPrismDirection,
		"os_v_prism_direction":    g.OsVPrismDirection,
		"od_dpd":                  g.OdDpd,
		"os_dpd":                  g.OsDpd,
	}
	if g.ExpirationDate != nil {
		m["expiration_date"] = g.ExpirationDate.Format("2006-01-02")
	} else {
		m["expiration_date"] = nil
	}
	return m
}

// internal/models/lab_ticket/lab_ticket_powers.go
package lab_ticket

import "fmt"

// ENUM wrappers (Postgres enums h_prism_direction_enum / v_prism_direction_enum)
type HPrismDirection string
type VPrismDirection string

// LabTicketPowers ↔ table: lab_ticket_powers
type LabTicketPowers struct {
	IDLabTicketPowers int64 `gorm:"column:id_lab_ticket_powers;primaryKey;autoIncrement" json:"id_lab_ticket_powers"`

	ODSph  *string `gorm:"column:od_sph;type:varchar(20)" json:"od_sph,omitempty"`
	OSSph  *string `gorm:"column:os_sph;type:varchar(20)" json:"os_sph,omitempty"`
	ODCyl  *string `gorm:"column:od_cyl;type:varchar(20)" json:"od_cyl,omitempty"`
	OSCyl  *string `gorm:"column:os_cyl;type:varchar(20)" json:"os_cyl,omitempty"`
	ODAxis *string `gorm:"column:od_axis;type:varchar(20)" json:"od_axis,omitempty"`
	OSAxis *string `gorm:"column:os_axis;type:varchar(20)" json:"os_axis,omitempty"`

	ODAdd *float64 `gorm:"column:od_add;type:numeric(5,2)" json:"od_add,omitempty"`
	OSAdd *float64 `gorm:"column:os_add;type:numeric(5,2)" json:"os_add,omitempty"`

	ODHPrism          *float64         `gorm:"column:od_h_prism;type:numeric(5,2)"              json:"od_h_prism,omitempty"`
	ODHPrismDirection *HPrismDirection `gorm:"column:od_h_prism_direction;type:h_prism_direction_enum" json:"od_h_prism_direction,omitempty"`
	OSHPrism          *float64         `gorm:"column:os_h_prism;type:numeric(5,2)"              json:"os_h_prism,omitempty"`
	OSHPrismDirection *HPrismDirection `gorm:"column:os_h_prism_direction;type:h_prism_direction_enum" json:"os_h_prism_direction,omitempty"`

	ODVPrism          *float64         `gorm:"column:od_v_prism;type:numeric(5,2)"              json:"od_v_prism,omitempty"`
	ODVPrismDirection *VPrismDirection `gorm:"column:od_v_prism_direction;type:v_prism_direction_enum" json:"od_v_prism_direction,omitempty"`
	OSVPrism          *float64         `gorm:"column:os_v_prism;type:numeric(5,2)"              json:"os_v_prism,omitempty"`
	OSVPrismDirection *VPrismDirection `gorm:"column:os_v_prism_direction;type:v_prism_direction_enum" json:"os_v_prism_direction,omitempty"`

	ODSegHD *string `gorm:"column:od_seg_hd;type:varchar(6)" json:"od_seg_hd,omitempty"`
	OSSegHD *string `gorm:"column:os_seg_hd;type:varchar(6)" json:"os_seg_hd,omitempty"`
	ODOC    *string `gorm:"column:od_oc;type:varchar(6)"     json:"od_oc,omitempty"`
	OSOC    *string `gorm:"column:os_oc;type:varchar(6)"     json:"os_oc,omitempty"`

	ODDT *string `gorm:"column:od_dt;type:varchar(6)" json:"od_dt,omitempty"`
	OSDT *string `gorm:"column:os_dt;type:varchar(6)" json:"os_dt,omitempty"`
	ODNR *string `gorm:"column:od_nr;type:varchar(6)" json:"od_nr,omitempty"`
	OSNR *string `gorm:"column:os_nr;type:varchar(6)" json:"os_nr,omitempty"`
	OUDT *string `gorm:"column:ou_dt;type:varchar(6)" json:"ou_dt,omitempty"`
	OUNR *string `gorm:"column:ou_nr;type:varchar(6)" json:"ou_nr,omitempty"`

	ODBC *string `gorm:"column:od_bc;type:varchar(6)" json:"od_bc,omitempty"`
	OSBC *string `gorm:"column:os_bc;type:varchar(6)" json:"os_bc,omitempty"`
}

func (LabTicketPowers) TableName() string { return "lab_ticket_powers" }

func (l *LabTicketPowers) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lab_ticket_powers": l.IDLabTicketPowers,
		"od_sph":               l.ODSph,
		"os_sph":               l.OSSph,
		"od_cyl":               l.ODCyl,
		"os_cyl":               l.OSCyl,
		"od_axis":              l.ODAxis,
		"os_axis":              l.OSAxis,
		"od_add":               l.ODAdd,
		"os_add":               l.OSAdd,

		"od_h_prism":           l.ODHPrism,
		"od_h_prism_direction": l.ODHPrismDirection,
		"os_h_prism":           l.OSHPrism,
		"os_h_prism_direction": l.OSHPrismDirection,

		"od_v_prism":           l.ODVPrism,
		"od_v_prism_direction": l.ODVPrismDirection,
		"os_v_prism":           l.OSVPrism,
		"os_v_prism_direction": l.OSVPrismDirection,

		"od_seg_hd": l.ODSegHD,
		"os_seg_hd": l.OSSegHD,
		"od_oc":     l.ODOC,
		"os_oc":     l.OSOC,

		"od_dt": l.ODDT,
		"os_dt": l.OSDT,
		"od_nr": l.ODNR,
		"os_nr": l.OSNR,
		"ou_dt": l.OUDT,
		"ou_nr": l.OUNR,

		"od_bc": l.ODBC,
		"os_bc": l.OSBC,
	}
}

func (l *LabTicketPowers) String() string {
	return fmt.Sprintf("<LabTicketPowers %d>", l.IDLabTicketPowers)
}

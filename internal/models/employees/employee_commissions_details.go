// internal/models/employees/employee_commissions_details.go
package employees

import "fmt"

// EmployeeCommissionsDetails ⇄ employee_commissions_details
type EmployeeCommissionsDetails struct {
	IDDetails    int      `gorm:"column:id_details;primaryKey"                                         json:"id_details"`
	PBKey        string   `gorm:"column:pb_key;type:commission_pb_key_enum;not null"                    json:"pb_key"`
	PercentValue float64  `gorm:"column:percent_value;type:double precision;not null"                   json:"percent_value"`
	SumMin       *float64 `gorm:"column:sum_min;type:double precision"                                  json:"sum_min,omitempty"`
	BrandID      *int     `gorm:"column:brand_id"                                                       json:"brand_id,omitempty"`
}

func (EmployeeCommissionsDetails) TableName() string { return "employee_commissions_details" }

func (d *EmployeeCommissionsDetails) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_details":    d.IDDetails,
		"pb_key":        d.PBKey,
		"percent_value": d.PercentValue,
		"sum_min":       d.SumMin,
		"brand_id":      d.BrandID,
	}
}

func (d *EmployeeCommissionsDetails) String() string {
	return fmt.Sprintf("<EmployeeCommissionsDetails %d>", d.IDDetails)
}

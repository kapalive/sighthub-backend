// internal/models/general/payroll_type.go
package general

import "fmt"

type PayrollType struct {
	IDPayrollType   int    `gorm:"column:id_payroll_type;primaryKey"           json:"id_payroll_type"`
	PayrollTypeName string `gorm:"column:payroll_type_name;type:varchar(100);not null" json:"payroll_type_name"`
}

func (PayrollType) TableName() string { return "payroll_type" }

func (p *PayrollType) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_payroll_type":   p.IDPayrollType,
		"payroll_type_name": p.PayrollTypeName,
	}
}

func (p *PayrollType) String() string {
	return fmt.Sprintf("<PayrollType %s>", p.PayrollTypeName)
}

// internal/models/employees/employee_commissions_details_relation.go
package employees

import "fmt"

// EmployeeCommissionsDetailsRelation ⇄ employee_commissions_details_relation
type EmployeeCommissionsDetailsRelation struct {
	RelationID            int `gorm:"column:relation_id;primaryKey"                                     json:"relation_id"`
	EmployeeCommissionsID int `gorm:"column:employee_commissions_id;not null"                           json:"employee_commissions_id"`
	DetailsID             int `gorm:"column:details_id;not null"                                        json:"details_id"`
}

func (EmployeeCommissionsDetailsRelation) TableName() string {
	return "employee_commissions_details_relation"
}

func (r *EmployeeCommissionsDetailsRelation) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"relation_id":             r.RelationID,
		"employee_commissions_id": r.EmployeeCommissionsID,
		"details_id":              r.DetailsID,
	}
}

func (r *EmployeeCommissionsDetailsRelation) String() string {
	return fmt.Sprintf("<EmployeeCommissionsDetailsRelation %d>", r.RelationID)
}

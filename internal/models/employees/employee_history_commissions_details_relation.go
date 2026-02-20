// internal/models/employees/employee_history_commissions_details_relation.go
package employees

import "fmt"

// EmployeeHistoryCommissionsDetailsRelation ⇄ employee_history_commissions_details_relation
type EmployeeHistoryCommissionsDetailsRelation struct {
	RelationID           int `gorm:"column:relation_id;primaryKey"                                     json:"relation_id"`
	HistoryCommissionsID int `gorm:"column:history_commissions_id;not null"                            json:"history_commissions_id"`
	DetailsID            int `gorm:"column:details_id;not null"                                        json:"details_id"`
}

func (EmployeeHistoryCommissionsDetailsRelation) TableName() string {
	return "employee_history_commissions_details_relation"
}

func (r *EmployeeHistoryCommissionsDetailsRelation) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"relation_id":            r.RelationID,
		"history_commissions_id": r.HistoryCommissionsID,
		"details_id":             r.DetailsID,
	}
}

func (r *EmployeeHistoryCommissionsDetailsRelation) String() string {
	return fmt.Sprintf("<EmployeeHistoryCommissionsDetailsRelation %d>", r.RelationID)
}

// internal/models/patients/client_balance.go
package patients

// ClientBalance ⇄ table: client_balance
// Хранит кредитный баланс пациента по локации.
type ClientBalance struct {
	IDClientBalance int64   `gorm:"column:id_client_balance;primaryKey;autoIncrement"                        json:"id_client_balance"`
	PatientID       int64   `gorm:"column:patient_id;not null;uniqueIndex:uix_patient_location"              json:"patient_id"`
	Credit          float64 `gorm:"column:credit;type:numeric(14,2);not null;default:0.00"                   json:"credit"`
	LocationID      int     `gorm:"column:location_id;not null;uniqueIndex:uix_patient_location"             json:"location_id"`
}

func (ClientBalance) TableName() string { return "client_balance" }

func (c *ClientBalance) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_client_balance": c.IDClientBalance,
		"patient_id":        c.PatientID,
		"credit":            c.Credit,
		"location_id":       c.LocationID,
	}
}

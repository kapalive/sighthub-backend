package marketing

import "time"

// GiftCard ⇄ table: gift_card
type GiftCard struct {
	IDGiftCard         int        `gorm:"column:id_gift_card;primaryKey;autoIncrement"                       json:"id_gift_card"`
	Code               string     `gorm:"column:code;type:varchar(50);not null"                              json:"code"`
	Nominal            string     `gorm:"column:nominal;type:numeric(10,2);not null"                         json:"nominal"`
	Balance            string     `gorm:"column:balance;type:numeric(10,2);not null"                         json:"balance"`
	ExpirationDate     *time.Time `gorm:"column:expiration_date;type:date"                                   json:"expiration_date,omitempty"`
	LocationID         int        `gorm:"column:location_id;not null;index"                                  json:"location_id"`
	InvoiceID          *int       `gorm:"column:invoice_id"                                                  json:"invoice_id,omitempty"`
	RecipientPatientID *int64     `gorm:"column:recipient_patient_id"                                        json:"recipient_patient_id,omitempty"`
	Status             string     `gorm:"column:status;type:varchar(50);not null;default:'active'"           json:"status"`
	CreatedAt          *time.Time `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"       json:"created_at,omitempty"`
}

func (GiftCard) TableName() string { return "gift_card" }

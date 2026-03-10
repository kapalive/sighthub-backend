// internal/models/lab_ticket/lab_ticket_invoice_item.go
package lab_ticket

// LabTicketInvoiceItem ↔ table: lab_ticket_invoice_item
// Связь между lab_ticket и invoice_item_sale.
type LabTicketInvoiceItem struct {
	IDLabTicketInvoiceItem int64    `gorm:"column:id_lab_ticket_invoice_item;primaryKey;autoIncrement"    json:"id_lab_ticket_invoice_item"`
	LabTicketID            int64    `gorm:"column:lab_ticket_id;not null;index;uniqueIndex:uix_lt_item"   json:"lab_ticket_id"`
	InvoiceItemID          int64    `gorm:"column:invoice_item_id;not null;index;uniqueIndex:uix_lt_item" json:"invoice_item_id"`
	CostOverride           *float64 `gorm:"column:cost_override;type:numeric(10,2)"                        json:"cost_override,omitempty"`
}

func (LabTicketInvoiceItem) TableName() string { return "lab_ticket_invoice_item" }

func (l *LabTicketInvoiceItem) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lab_ticket_invoice_item": l.IDLabTicketInvoiceItem,
		"lab_ticket_id":              l.LabTicketID,
		"invoice_item_id":            l.InvoiceItemID,
		"cost_override":              l.CostOverride,
	}
}

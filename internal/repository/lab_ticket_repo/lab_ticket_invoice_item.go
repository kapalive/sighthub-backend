// internal/repository/lab_ticket_repo/lab_ticket_invoice_item.go
package lab_ticket_repo

import (
	"errors"
	"gorm.io/gorm"
	lt "sighthub-backend/internal/models/lab_ticket"
)

type LabTicketInvoiceItemRepo struct{ DB *gorm.DB }

func NewLabTicketInvoiceItemRepo(db *gorm.DB) *LabTicketInvoiceItemRepo {
	return &LabTicketInvoiceItemRepo{DB: db}
}

// GetByLabTicketID возвращает связанные invoice_item_sale записи.
func (r *LabTicketInvoiceItemRepo) GetByLabTicketID(labTicketID int64) ([]lt.LabTicketInvoiceItem, error) {
	var rows []lt.LabTicketInvoiceItem
	return rows, r.DB.Where("lab_ticket_id = ?", labTicketID).Find(&rows).Error
}

// Link привязывает invoice_item к lab ticket.
func (r *LabTicketInvoiceItemRepo) Link(labTicketID, invoiceItemID int64, costOverride *float64) error {
	item := &lt.LabTicketInvoiceItem{
		LabTicketID:   labTicketID,
		InvoiceItemID: invoiceItemID,
		CostOverride:  costOverride,
	}
	return r.DB.Create(item).Error
}

// Unlink отвязывает invoice_item от lab ticket.
func (r *LabTicketInvoiceItemRepo) Unlink(labTicketID, invoiceItemID int64) error {
	return r.DB.
		Where("lab_ticket_id = ? AND invoice_item_id = ?", labTicketID, invoiceItemID).
		Delete(&lt.LabTicketInvoiceItem{}).Error
}

// UnlinkAllForTicket удаляет все связи для тикета.
func (r *LabTicketInvoiceItemRepo) UnlinkAllForTicket(labTicketID int64) error {
	return r.DB.Where("lab_ticket_id = ?", labTicketID).Delete(&lt.LabTicketInvoiceItem{}).Error
}

func (r *LabTicketInvoiceItemRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

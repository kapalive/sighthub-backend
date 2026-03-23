package ticket_service

import (
	"fmt"

	invoiceModel "sighthub-backend/internal/models/invoices"
	"sighthub-backend/internal/models/lenses"
	serviceModel "sighthub-backend/internal/models/service"
)

// ─── Ticket Lens Options (items from invoice) ────────────────────────────────

type TicketLensOption struct {
	ItemName    string  `json:"item_name"`
	Description *string `json:"description"`
}

// GetTicketLensOptions returns lenses, treatments and additional services
// that are already in the ticket's invoice.
func (s *Service) GetTicketLensOptions(ticketID int64) ([]TicketLensOption, error) {
	var ticket struct{ InvoiceID int64 }
	if err := s.db.Table("lab_ticket").
		Select("invoice_id").
		Where("id_lab_ticket = ?", ticketID).
		Scan(&ticket).Error; err != nil || ticket.InvoiceID == 0 {
		return nil, fmt.Errorf("ticket not found")
	}

	var items []invoiceModel.InvoiceItemSale
	s.db.Where("invoice_id = ? AND item_type IN ?", ticket.InvoiceID,
		[]string{"Treatment", "Add service"}).
		Find(&items)

	result := make([]TicketLensOption, 0, len(items))
	for _, item := range items {
		opt := TicketLensOption{
			Description: &item.Description,
		}

		if item.ItemID != nil {
			switch item.ItemType {
			case "Treatment":
				var t lenses.LensTreatments
				if s.db.First(&t, *item.ItemID).Error == nil {
					opt.ItemName = t.ItemNbr
				}
			case "Add service":
				var a serviceModel.AdditionalService
				if s.db.First(&a, *item.ItemID).Error == nil && a.ItemNumber != nil {
					opt.ItemName = *a.ItemNumber
				}
			}
		}

		if opt.ItemName == "" {
			opt.ItemName = item.Description
		}

		result = append(result, opt)
	}

	return result, nil
}

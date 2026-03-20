package invoice_service

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/invoices"
	"sighthub-backend/internal/models/location"
)

// ─── Update Invoice Status ────────────────────────────────────────────────────

func (s *Service) UpdateInvoiceStatus(invoiceID int64, statusInvoiceID int) (map[string]interface{}, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: invoice not found", ErrNotFound)
	}

	var si invoices.StatusInvoice
	if err := s.db.First(&si, statusInvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: status not found", ErrNotFound)
	}

	statusID := int64(statusInvoiceID); inv.StatusInvoiceID = &statusID
	s.db.Save(&inv)

	return map[string]interface{}{
		"message":           "Status updated",
		"status_invoice_id": statusInvoiceID,
		"status_invoice":    si.StatusInvoiceValue,
	}, nil
}

// ─── Get Invoice Statuses ─────────────────────────────────────────────────────

func (s *Service) GetInvoiceStatuses() ([]map[string]interface{}, error) {
	var statuses []invoices.StatusInvoice
	s.db.Where("id_status_invoice IN ?", []int{24, 25, 26, 27}).
		Order("id_status_invoice").Find(&statuses)

	var result []map[string]interface{}
	for _, si := range statuses {
		result = append(result, map[string]interface{}{
			"status_invoice_id": si.IDStatusInvoice,
			"status_invoice":    si.StatusInvoiceValue,
			"icon":              si.Icon,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

// ─── Get Patient Returns ──────────────────────────────────────────────────────

func (s *Service) GetPatientReturns(el *EmpLocation, dateFrom, dateTo *time.Time) ([]map[string]interface{}, error) {
	now := time.Now()
	df := now.AddDate(0, 0, -30)
	dt := now
	if dateFrom != nil {
		df = *dateFrom
	}
	if dateTo != nil {
		dt = *dateTo
	}

	locID := int64(el.Location.IDLocation)

	var returns []invoices.ReturnInvoice
	if err := s.db.
		Joins("JOIN invoice ON invoice.id_invoice = return_invoices.invoice_id").
		Where("invoice.location_id = ?", locID).
		Where("return_invoices.return_date >= ? AND return_invoices.return_date <= ?", df, dt).
		Find(&returns).Error; err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for _, r := range returns {
		var inv invoices.Invoice
		if err := s.db.Preload("Patient").First(&inv, r.InvoiceID).Error; err != nil {
			continue
		}
		var patientID *int64
		var patientName *string
		if inv.Patient != nil {
			id := int64(inv.Patient.IDPatient)
			patientID = &id
			name := fmt.Sprintf("%s %s", inv.Patient.FirstName, inv.Patient.LastName)
			patientName = &name
		}

		var returnDate *string
		if !r.ReturnDate.IsZero() {
			d := r.ReturnDate.Format("2006-01-02")
			returnDate = &d
		}

		results = append(results, map[string]interface{}{
			"return_id":     r.ReturnID,
			"invoice_id":    r.InvoiceID,
			"patient_id":    patientID,
			"patient_name":  patientName,
			"status":        r.Status,
			"return_reason": r.ReturnReason,
			"return_date":   returnDate,
			"return_amount": fmtFloat(r.ReturnAmount),
		})
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, nil
}

// ─── Get Transfer Locations ───────────────────────────────────────────────────

func (s *Service) GetTransferLocations(el *EmpLocation, transferType string) ([]map[string]interface{}, error) {
	if transferType != "local" && transferType != "foreign" {
		return nil, fmt.Errorf("%w: parameter 'type' must be 'local' or 'foreign'", ErrBadRequest)
	}

	var locs []location.Location

	if transferType == "local" {
		// Same store/warehouse group, excluding current location (like Python)
		var orConds []string
		var orArgs []interface{}
		if el.Location.StoreID != 0 {
			orConds = append(orConds, "store_id = ?")
			orArgs = append(orArgs, el.Location.StoreID)
		}
		if el.Location.WarehouseID != nil {
			orConds = append(orConds, "warehouse_id = ?")
			orArgs = append(orArgs, *el.Location.WarehouseID)
		}
		if len(orConds) == 0 {
			return []map[string]interface{}{}, nil
		}
		orSQL := "(" + strings.Join(orConds, " OR ") + ") AND id_location != ?"
		orArgs = append(orArgs, el.Location.IDLocation)
		if err := s.db.Where(orSQL, orArgs...).Find(&locs).Error; err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
	} else {
		s.db.Where("can_receive_items = true AND store_id != ? AND id_location != ?",
			el.Location.StoreID, el.Location.IDLocation).Find(&locs)
	}

	var result []map[string]interface{}
	for _, loc := range locs {
		result = append(result, map[string]interface{}{
			"location_id": loc.IDLocation,
			"location":    loc.FullName,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

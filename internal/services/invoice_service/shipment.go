package invoice_service

import (
	"fmt"
	"time"

	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/invoices"
	"sighthub-backend/internal/models/location"
)

// ─── Create Shipment ──────────────────────────────────────────────────────────

type CreateShipmentRequest struct {
	VendorID   int64   `json:"vendor_id"`
	LocationID int64   `json:"location_id"`
	BrandID    *int64  `json:"brand_id"`
	QtyOk      int     `json:"qty_ok"`
	QtyHold    int     `json:"qty_hold"`
	QtyShort   int     `json:"qty_short"`
	QtyOver    int     `json:"qty_over"`
	Cost       float64 `json:"cost"`
	Notes      *string `json:"notes"`
}

func (s *Service) CreateShipment(el *EmpLocation, req CreateShipmentRequest) (map[string]interface{}, error) {
	if req.VendorID == 0 || req.LocationID == 0 {
		return nil, fmt.Errorf("%w: vendor_id and location_id are required", ErrBadRequest)
	}

	empID := int64(el.Employee.IDEmployee)
	status := "Received"
	shipment := invModel.Shipment{
		VendorID:          req.VendorID,
		LocationID:        req.LocationID,
		QtyOk:             req.QtyOk,
		QtyHold:           req.QtyHold,
		QtyShort:          req.QtyShort,
		QtyOver:           req.QtyOver,
		Cost:              req.Cost,
		EmployeeIDPrepBy:  empID,
		EmployeeIDCreated: empID,
		DateReceived:      time.Now(),
		Status:            &status,
		Notes:             req.Notes,
	}
	if req.BrandID != nil {
		shipment.BrandID = *req.BrandID
	}

	if err := s.db.Create(&shipment).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":     "Shipment created successfully",
		"shipment_id": shipment.IDShipment,
	}, nil
}

// ─── Get Shipment ─────────────────────────────────────────────────────────────

func (s *Service) GetShipment(shipmentID int64) (map[string]interface{}, error) {
	var sh invModel.Shipment
	if err := s.db.First(&sh, shipmentID).Error; err != nil {
		return nil, fmt.Errorf("%w: shipment not found", ErrNotFound)
	}

	var dr *string
	if !sh.DateReceived.IsZero() {
		d := sh.DateReceived.Format(time.RFC3339)
		dr = &d
	}

	return map[string]interface{}{
		"id_shipment":  sh.IDShipment,
		"vendor_id":    sh.VendorID,
		"location_id":  sh.LocationID,
		"brand_id":     sh.BrandID,
		"qty_ok":       sh.QtyOk,
		"qty_hold":     sh.QtyHold,
		"qty_short":    sh.QtyShort,
		"qty_over":     sh.QtyOver,
		"cost":         fmtFloat(sh.Cost),
		"date_received": dr,
		"status":       sh.Status,
		"notes":        sh.Notes,
	}, nil
}

// ─── Update Shipment ──────────────────────────────────────────────────────────

type UpdateShipmentRequest struct {
	QtyOk    *int     `json:"qty_ok"`
	QtyHold  *int     `json:"qty_hold"`
	QtyShort *int     `json:"qty_short"`
	QtyOver  *int     `json:"qty_over"`
	Cost     *float64 `json:"cost"`
	Notes    *string  `json:"notes"`
}

func (s *Service) UpdateShipment(shipmentID int64, req UpdateShipmentRequest) (map[string]interface{}, error) {
	var sh invModel.Shipment
	if err := s.db.First(&sh, shipmentID).Error; err != nil {
		return nil, fmt.Errorf("%w: shipment not found", ErrNotFound)
	}

	if req.QtyOk != nil {
		sh.QtyOk = *req.QtyOk
	}
	if req.QtyHold != nil {
		sh.QtyHold = *req.QtyHold
	}
	if req.QtyShort != nil {
		sh.QtyShort = *req.QtyShort
	}
	if req.QtyOver != nil {
		sh.QtyOver = *req.QtyOver
	}
	if req.Cost != nil {
		sh.Cost = *req.Cost
	}
	if req.Notes != nil {
		sh.Notes = req.Notes
	}

	s.db.Save(&sh)
	return map[string]interface{}{"message": "Shipment updated successfully"}, nil
}

// ─── List Shipments ───────────────────────────────────────────────────────────

func (s *Service) GetShipments() ([]map[string]interface{}, error) {
	var shipments []invModel.Shipment
	s.db.Find(&shipments)

	var result []map[string]interface{}
	for _, sh := range shipments {
		var dr *string
		if !sh.DateReceived.IsZero() {
			d := sh.DateReceived.Format(time.RFC3339)
			dr = &d
		}
		result = append(result, map[string]interface{}{
			"id_shipment":  sh.IDShipment,
			"vendor_id":    sh.VendorID,
			"location_id":  sh.LocationID,
			"qty_ok":       sh.QtyOk,
			"cost":         fmtFloat(sh.Cost),
			"date_received": dr,
			"status":       sh.Status,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

// ─── Transfers ────────────────────────────────────────────────────────────────

type TransferFilter struct {
	Type     string // "local", "foreign", or ""
	DateFrom *time.Time
	DateTo   *time.Time
}

func (s *Service) GetTransfers(el *EmpLocation, f TransferFilter) ([]map[string]interface{}, error) {
	now := time.Now()
	dateFrom := now.AddDate(0, 0, -30)
	dateTo := now
	if f.DateFrom != nil {
		dateFrom = *f.DateFrom
	}
	if f.DateTo != nil {
		dateTo = *f.DateTo
	}

	locID := int64(el.Location.IDLocation)

	q := s.db.Model(&invoices.Invoice{}).
		Where("date_create BETWEEN ? AND ?", dateFrom, dateTo).
		Where("number_invoice NOT LIKE 'V%' AND number_invoice NOT LIKE 'S%'")

	switch f.Type {
	case "":
		q = q.Where("location_id = ?", locID)
	case "local":
		whID := 0
		if el.Location.WarehouseID != nil {
			whID = *el.Location.WarehouseID
		}
		q = q.Where(
			"(location_id = ? OR to_location_id = ?) AND ((location_id = ? AND to_location_id = ?) OR (to_location_id = ? AND location_id = ?))",
			locID, locID, locID, whID, locID, whID,
		)
	case "foreign":
		whID := 0
		if el.Location.WarehouseID != nil {
			whID = *el.Location.WarehouseID
		}
		q = q.Where("(to_location_id = ? OR location_id = ?)", locID, locID).
			Where("NOT ((location_id = ? AND to_location_id = ?) OR (to_location_id = ? AND location_id = ?))",
				locID, whID, locID, whID)
	}

	var invs []invoices.Invoice
	q.Order("date_create DESC").Find(&invs)

	var result []map[string]interface{}
	for _, inv := range invs {
		var fromLoc, toLoc location.Location
		fromName := "Unknown"
		toName := "Unknown"
		if err := s.db.First(&fromLoc, inv.LocationID).Error; err == nil {
			fromName = fromLoc.FullName
		}
		if inv.ToLocationID != nil {
			if err := s.db.First(&toLoc, *inv.ToLocationID).Error; err == nil {
				toName = toLoc.FullName
			}
		}

		statusName := "Unknown"
		var si invoices.StatusInvoice
		if err := s.db.First(&si, inv.StatusInvoiceID).Error; err == nil {
			statusName = si.StatusInvoiceValue
		}

		result = append(result, map[string]interface{}{
			"invoice_id":        inv.IDInvoice,
			"invoice_number":    inv.NumberInvoice,
			"from_location":     fromName,
			"to_location":       toName,
			"date_create":       inv.DateCreate.Format(time.RFC3339),
			"status_invoice_id": inv.StatusInvoiceID,
			"status_invoice":    statusName,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

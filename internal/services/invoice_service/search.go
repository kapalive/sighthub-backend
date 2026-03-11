package invoice_service

import (
	"fmt"
	"strconv"

	"sighthub-backend/internal/models/invoices"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/vendors"
)

// ─── Search Invoices ──────────────────────────────────────────────────────────

type SearchResult struct {
	IDInvoice     int64  `json:"id_invoice"`
	NumberInvoice string `json:"number_invoice"`
	RedirectURL   string `json:"redirect_url"`
}

func (s *Service) SearchInvoice(el *EmpLocation, query string) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("%w: invoice number not provided", ErrBadRequest)
	}

	pattern := "%" + query + "%"
	var invs []invoices.Invoice

	_, err := strconv.ParseInt(query, 10, 64)
	if err == nil {
		// Numeric: search by id and number
		s.db.Where("CAST(id_invoice AS TEXT) ILIKE ? OR number_invoice ILIKE ?", pattern, pattern).Find(&invs)
	} else {
		s.db.Where("number_invoice ILIKE ?", pattern).Find(&invs)
	}

	if len(invs) == 0 {
		return nil, fmt.Errorf("%w: no invoices found", ErrNotFound)
	}

	locID := int64(el.Location.IDLocation)
	var accessible []SearchResult

	for _, inv := range invs {
		switch {
		case len(inv.NumberInvoice) > 0 && inv.NumberInvoice[0] == 'V':
			if inv.LocationID == locID {
				accessible = append(accessible, SearchResult{
					IDInvoice:     inv.IDInvoice,
					NumberInvoice: inv.NumberInvoice,
					RedirectURL:   fmt.Sprintf("/receipts/vendor/invoice/%d/%d/*", inv.IDInvoice, inv.VendorID),
				})
			}
		case len(inv.NumberInvoice) > 0 && inv.NumberInvoice[0] == 'I':
			if inv.LocationID == locID {
				accessible = append(accessible, SearchResult{
					IDInvoice:     inv.IDInvoice,
					NumberInvoice: inv.NumberInvoice,
					RedirectURL:   fmt.Sprintf("/transfers/%d", inv.IDInvoice),
				})
			} else if inv.ToLocationID != nil && *inv.ToLocationID == locID {
				accessible = append(accessible, SearchResult{
					IDInvoice:     inv.IDInvoice,
					NumberInvoice: inv.NumberInvoice,
					RedirectURL:   fmt.Sprintf("/receipts/location/invoice/%d", inv.IDInvoice),
				})
			}
		case len(inv.NumberInvoice) > 0 && inv.NumberInvoice[0] == 'S':
			if inv.LocationID == locID {
				accessible = append(accessible, SearchResult{
					IDInvoice:     inv.IDInvoice,
					NumberInvoice: inv.NumberInvoice,
					RedirectURL:   fmt.Sprintf("/patient/%d/invoice/%d", inv.PatientID, inv.IDInvoice),
				})
			}
		}
	}

	if len(accessible) == 0 {
		return nil, fmt.Errorf("%w: no accessible invoices found", ErrForbidden)
	}
	return accessible, nil
}

// ─── Vendors List ─────────────────────────────────────────────────────────────

func (s *Service) GetVendors() ([]map[string]interface{}, error) {
	var vs []vendors.Vendor
	s.db.Order("vendor_name ASC").Find(&vs)

	var result []map[string]interface{}
	for _, v := range vs {
		result = append(result, map[string]interface{}{
			"vendor_id":   v.IDVendor,
			"vendor_name": v.VendorName,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

// ─── Locations List ───────────────────────────────────────────────────────────

func (s *Service) GetReceiptLocations() ([]map[string]interface{}, error) {
	var locs []location.Location
	s.db.Where("can_receive_items = true AND store_active = true").Find(&locs)

	var result []map[string]interface{}
	for _, loc := range locs {
		result = append(result, map[string]interface{}{
			"location_id":   loc.IDLocation,
			"location_name": loc.FullName,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

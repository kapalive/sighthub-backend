package invoice_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	invoiceSvc "sighthub-backend/internal/services/invoice_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *invoiceSvc.Service
}

func New(svc *invoiceSvc.Service) *Handler { return &Handler{svc: svc} }

// ─── Invoice CRUD ─────────────────────────────────────────────────────────────

// POST /api/invoice/create
func (h *Handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req invoiceSvc.CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.CreateInvoice(el, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// PUT /api/invoice/update/{invoice_id}
func (h *Handler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var body struct {
		DateCreate *string            `json:"date_create"`
		Discount   *float64           `json:"discount"`
		Items      []invoiceSvc.ItemIn `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.UpdateInvoice(el, id, body.DateCreate, body.Discount, body.Items)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/view/{invoice_id}
func (h *Handler) ViewInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	result, err := h.svc.ViewInvoice(el, id)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/view-item/{invoice_id}
func (h *Handler) ViewInvoiceItem(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	result, err := h.svc.ViewInvoiceItem(el, id)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// DELETE /api/invoice/delete-item
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	sku := r.URL.Query().Get("sku")
	invoiceIDStr := r.URL.Query().Get("invoice_id")
	inventoryIDStr := r.URL.Query().Get("inventory_id")

	invoiceID, err := strconv.ParseInt(invoiceIDStr, 10, 64)
	if err != nil || invoiceID == 0 {
		jsonError(w, "invoice_id is required", http.StatusBadRequest)
		return
	}

	var inventoryID *int64
	if inventoryIDStr != "" {
		id, err := strconv.ParseInt(inventoryIDStr, 10, 64)
		if err == nil {
			inventoryID = &id
		}
	}

	if sku == "" && inventoryID == nil {
		jsonError(w, "sku or inventory_id is required", http.StatusBadRequest)
		return
	}

	result, err := h.svc.DeleteItem(invoiceID, sku, inventoryID)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// DELETE /api/invoice/delete/{invoice_id}
func (h *Handler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.DeleteInvoice(id)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── Receipt ──────────────────────────────────────────────────────────────────

// GET /api/invoice/receipt
func (h *Handler) GetReceipts(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	f := invoiceSvc.ReceiptFilter{
		InvoiceType: r.URL.Query().Get("invoice_type"),
	}
	if v := r.URL.Query().Get("vendor_id"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		f.VendorID = &id
	}
	if v := r.URL.Query().Get("from_location_id"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		f.FromLocationID = &id
	}
	if v := r.URL.Query().Get("date_from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			jsonError(w, "invalid date_from", http.StatusBadRequest)
			return
		}
		f.DateFrom = &t
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			jsonError(w, "invalid date_to", http.StatusBadRequest)
			return
		}
		// end of day
		t = t.Add(24*time.Hour - time.Second)
		f.DateTo = &t
	}

	result, err := h.svc.GetReceipts(el, f)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/receipt/{invoice_id}
func (h *Handler) GetReceipt(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetReceipt(id)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// PUT /api/invoice/receipt/confirm
func (h *Handler) ConfirmReceipt(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req invoiceSvc.ConfirmReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.ConfirmReceipt(el, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /api/invoice/receipt/{invoice_id}/pay
func (h *Handler) PayTransfer(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	result, err := h.svc.PayTransfer(el, id)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── Vendor Invoices ──────────────────────────────────────────────────────────

// POST /api/invoice/vendor_invoices/create
func (h *Handler) CreateVendorInvoice(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req invoiceSvc.CreateVendorInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.CreateVendorInvoice(el, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// PUT /api/invoice/vendor_invoices/{id}
func (h *Handler) UpdateVendorInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req invoiceSvc.UpdateVendorInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.UpdateVendorInvoice(el, id, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/vendor_invoice/{id}
func (h *Handler) GetVendorInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	result, err := h.svc.GetVendorInvoice(el, id)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/vendor-contacts?vendor_id=X
func (h *Handler) GetVendorContacts(w http.ResponseWriter, r *http.Request) {
	vid, err := strconv.Atoi(r.URL.Query().Get("vendor_id"))
	if err != nil || vid == 0 {
		jsonError(w, "vendor_id is required", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetVendorContacts(vid)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/location-contacts?location_id=X
func (h *Handler) GetLocationContacts(w http.ResponseWriter, r *http.Request) {
	lid, err := strconv.Atoi(r.URL.Query().Get("location_id"))
	if err != nil || lid == 0 {
		jsonError(w, "location_id is required", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetLocationContacts(lid)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── Shipment ─────────────────────────────────────────────────────────────────

// POST /api/invoice/shipment/create
func (h *Handler) CreateShipment(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req invoiceSvc.CreateShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.CreateShipment(el, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// GET /api/invoice/shipment/{shipment_id}
func (h *Handler) GetShipment(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "shipment_id")
	if err != nil {
		jsonError(w, "invalid shipment_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetShipment(id)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// PUT /api/invoice/shipment/update/{shipment_id}
func (h *Handler) UpdateShipment(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "shipment_id")
	if err != nil {
		jsonError(w, "invalid shipment_id", http.StatusBadRequest)
		return
	}
	var req invoiceSvc.UpdateShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.UpdateShipment(id, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/shipments
func (h *Handler) GetShipments(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetShipments()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/transfers
func (h *Handler) GetTransfers(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonResponse(w, []interface{}{}, http.StatusOK)
		return
	}

	f := invoiceSvc.TransferFilter{Type: r.URL.Query().Get("type")}
	if v := r.URL.Query().Get("date_from"); v != "" {
		t, _ := time.Parse("2006-01-02", v)
		f.DateFrom = &t
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		t, _ := time.Parse("2006-01-02", v)
		f.DateTo = &t
	}

	result, err := h.svc.GetTransfers(el, f)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── Return Invoices ──────────────────────────────────────────────────────────

// GET /api/invoice/return_invoices
func (h *Handler) GetReturnInvoices(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonResponse(w, []interface{}{}, http.StatusOK)
		return
	}

	f := invoiceSvc.ReturnInvoiceFilter{
		GroupByVendor: r.URL.Query().Get("group") == "vendor",
	}
	if v := r.URL.Query().Get("vendor_id"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		f.VendorID = &id
	}
	if v := r.URL.Query().Get("date_from"); v != "" {
		t, _ := time.Parse("2006-01-02", v)
		f.DateFrom = &t
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		t, _ := time.Parse("2006-01-02", v)
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		f.DateTo = &t
	}

	result, err := h.svc.GetReturnInvoices(el, f)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /api/invoice/return_invoice
func (h *Handler) CreateReturnInvoice(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req invoiceSvc.CreateReturnInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.CreateReturnInvoice(el, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// PUT /api/invoice/return_invoice/{id}
func (h *Handler) UpdateReturnInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req invoiceSvc.UpdateReturnInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.UpdateReturnInvoice(el, id, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/return_invoice/{id}
func (h *Handler) GetReturnInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetReturnInvoice(id)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// DELETE /api/invoice/return_invoice/{id}
func (h *Handler) DeleteReturnInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteReturnInvoice(id); err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, map[string]string{"message": "Return invoice and associated items deleted successfully"}, http.StatusOK)
}

// GET /api/invoice/return_invoice/shipping-services
func (h *Handler) GetShippingServices(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetShippingServices()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/return_invoice/payment-methods
func (h *Handler) GetReturnPaymentMethods(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetReturnPaymentMethods()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /api/invoice/return_invoice/{id}/payment
func (h *Handler) AddReturnPayment(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req invoiceSvc.AddReturnPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.AddReturnPayment(el, id, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// GET /api/invoice/return_invoice/{id}/payments
func (h *Handler) GetReturnPayments(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetReturnPayments(id)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// DELETE /api/invoice/return_invoice/{id}/payment/{payment_id}
func (h *Handler) DeleteReturnPayment(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	paymentID, err := pathInt64(r, "payment_id")
	if err != nil {
		jsonError(w, "invalid payment_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.DeleteReturnPayment(id, paymentID)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── Search / Lists ───────────────────────────────────────────────────────────

// GET /api/invoice/search?number_invoice=X
func (h *Handler) SearchInvoice(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	query := r.URL.Query().Get("number_invoice")
	results, err := h.svc.SearchInvoice(el, query)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, map[string]interface{}{"invoices": results}, http.StatusOK)
}

// GET /api/invoice/vendors
func (h *Handler) GetVendors(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetVendors()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /api/invoice/locations
func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetReceiptLocations()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func pathInt64(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)[key], 10, 64)
}

func httpStatus(err error) int {
	switch {
	case errors.Is(err, invoiceSvc.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, invoiceSvc.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, invoiceSvc.ErrBadRequest):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

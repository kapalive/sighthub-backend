package accounting_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	accSvc "sighthub-backend/internal/services/accounting_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *accSvc.Service }

func New(svc *accSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func errorStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"),
		strings.Contains(msg, "mismatch"), strings.Contains(msg, "cannot"):
		return http.StatusBadRequest
	case strings.Contains(msg, "different vendor accounts"), strings.Contains(msg, "already used"):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func pathInt(r *http.Request, key string) (int, error) {
	return strconv.Atoi(mux.Vars(r)[key])
}

func pathInt64(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)[key], 10, 64)
}

// GET /vendors
func (h *Handler) GetVendors(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetVendors()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /:vendor_id/quickbooks-header
func (h *Handler) GetVendorQuickbooksHeader(w http.ResponseWriter, r *http.Request) {
	vendorID, err := pathInt(r, "vendor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid vendor_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.GetVendorQuickbooksHeader(username, vendorID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /payments_methods
func (h *Handler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetPaymentMethods()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /stores
func (h *Handler) GetStores(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetStores()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /locations
func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetLocations()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /vendor-invoices/:vendor_id
func (h *Handler) GetVendorInvoicesList(w http.ResponseWriter, r *http.Request) {
	vendorID, err := pathInt(r, "vendor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid vendor_id"})
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.GetVendorInvoicesList(username, vendorID, page, perPage)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /invoices/:vendor_id
func (h *Handler) GetInvoicesByVendor(w http.ResponseWriter, r *http.Request) {
	vendorID, err := pathInt(r, "vendor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid vendor_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.GetInvoicesByVendor(username, vendorID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// POST /vendor-bills
func (h *Handler) CreateVendorBill(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.CreateVendorBill(username, data)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

// GET /transactions/:vendor_id
func (h *Handler) GetTransactionsByVendor(w http.ResponseWriter, r *http.Request) {
	vendorID, err := pathInt(r, "vendor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid vendor_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.GetTransactionsByVendor(username, vendorID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// POST /add_payment
func (h *Handler) AddPaymentToVendor(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.AddPaymentToVendor(username, data)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

// GET /vendors-balances
func (h *Handler) GetVendorsBalances(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.GetVendorsBalances(username)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /vendors/:vendor_id/account-number
func (h *Handler) ListVendorLocationAccounts(w http.ResponseWriter, r *http.Request) {
	vendorID, err := pathInt(r, "vendor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid vendor_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	statusFilter := r.URL.Query().Get("status")
	var isActive *bool
	if v := r.URL.Query().Get("is_active"); v != "" {
		b := v == "1" || v == "true" || v == "yes"
		isActive = &b
	}

	result, err := h.svc.ListVendorLocationAccounts(username, vendorID, statusFilter, isActive)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// PUT /vendors/:vendor_id/account-number/:acc_id
func (h *Handler) UpdateVendorAccountNumber(w http.ResponseWriter, r *http.Request) {
	vendorID, err := pathInt(r, "vendor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid vendor_id"})
		return
	}
	accID, err := pathInt64(r, "acc_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid acc_id"})
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.UpdateVendorAccountNumber(username, vendorID, accID, data)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// DELETE /vendors/:vendor_id/account-number/:acc_id
func (h *Handler) DeleteVendorAccountNumber(w http.ResponseWriter, r *http.Request) {
	vendorID, err := pathInt(r, "vendor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid vendor_id"})
		return
	}
	accID, err := pathInt64(r, "acc_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid acc_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.DeleteVendorAccountNumber(username, vendorID, accID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// POST /vendors/:vendor_id/account-number
func (h *Handler) CreateVendorAccountNumber(w http.ResponseWriter, r *http.Request) {
	vendorID, err := pathInt(r, "vendor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid vendor_id"})
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.CreateVendorAccountNumber(username, vendorID, data)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

// GET /notify/terms
func (h *Handler) GetTermsNotifyList(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days == 0 {
		days = 5
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 200
	}
	var vendorID *int
	if v := r.URL.Query().Get("vendor_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			vendorID = &id
		}
	}
	result, err := h.svc.GetTermsNotifyList(username, days, vendorID, limit)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /return-to-vendor-invoices/:vendor_id
func (h *Handler) GetReturnToVendorInvoices(w http.ResponseWriter, r *http.Request) {
	vendorID, err := pathInt(r, "vendor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid vendor_id"})
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	result, err := h.svc.GetReturnToVendorInvoices(vendorID, page, perPage)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// PUT /return-to-vendor-invoices/:rtv_id/credit
func (h *Handler) UpdateReturnToVendorCredit(w http.ResponseWriter, r *http.Request) {
	rtvID, err := pathInt64(r, "rtv_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid rtv_id"})
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	creditRaw, ok := data["credit_amount"]
	if !ok || creditRaw == nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "credit_amount is required"})
		return
	}
	creditAmount, ok := creditRaw.(float64)
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "credit_amount must be a number"})
		return
	}
	result, err := h.svc.UpdateReturnToVendorCredit(rtvID, creditAmount)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /ledger/:vendor_id
func (h *Handler) GetVendorLedger(w http.ResponseWriter, r *http.Request) {
	vendorID, err := pathInt(r, "vendor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid vendor_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.GetVendorLedger(username, vendorID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /ledger/entry
func (h *Handler) GetLedgerEntry(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	entryType := strings.TrimSpace(r.URL.Query().Get("type"))
	entryIDStr := r.URL.Query().Get("id")
	if entryType == "" || entryIDStr == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "type and id are required"})
		return
	}
	entryID, err := strconv.ParseInt(entryIDStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	result, err := h.svc.GetLedgerEntry(username, entryType, entryID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

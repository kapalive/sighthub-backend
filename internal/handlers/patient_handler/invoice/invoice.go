package invoice

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	invSvc "sighthub-backend/internal/services/patient_service/invoice"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *invSvc.Service
}

func New(svc *invSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}

func parseInvoiceID(r *http.Request) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)["invoice_id"], 10, 64)
}

// GET /invoice
func (h *Handler) GetInvoiceList(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id_patient")
	if idStr == "" {
		jsonError(w, "id_patient is required", http.StatusBadRequest)
		return
	}
	patientID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid id_patient", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.GetInvoiceList(username, patientID)
	if err != nil {
		switch err.Error() {
		case "employee or location not found", "patient not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		default:
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /invoice
func (h *Handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PatientID int64 `json:"patient_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.PatientID == 0 {
		jsonError(w, "Patient ID is required", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.CreateInvoice(username, body.PatientID)
	if err != nil {
		switch err.Error() {
		case "employee or location not found", "patient not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		default:
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// POST /invoice/{invoice_id}/remake
func (h *Handler) CreateRemakeInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.CreateRemakeInvoice(username, invoiceID)
	if err != nil {
		switch err.Error() {
		case "employee or location not found", "invoice not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		case "remake can only be created at the location where the original invoice was created":
			jsonError(w, err.Error(), http.StatusForbidden)
		default:
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// GET /payment-methods
func (h *Handler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetPaymentMethods()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /invoice/{invoice_id}
func (h *Handler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	groupByFrame := r.URL.Query().Get("group_by_frame") == "true"
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.GetInvoice(username, invoiceID, groupByFrame)
	if err != nil {
		switch err.Error() {
		case "employee or location not found", "invoice not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		default:
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /invoice/{invoice_id}/html
func (h *Handler) GetInvoiceHTML(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	ctx, err := h.svc.BuildInvoiceHTMLContext(invoiceID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, ctx, http.StatusOK)
}

// GET /invoice/{invoice_id}/print
func (h *Handler) PrintInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	htmlContent, err := h.svc.RenderInvoiceHTML(invoiceID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlContent)) //nolint:errcheck
}

// GET /invoice/{invoice_id}/pdf
func (h *Handler) GetInvoicePDF(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	pdfBytes, invoiceNumber, err := h.svc.GenerateInvoicePDF(invoiceID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=\"invoice-"+invoiceNumber+".pdf\"")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes) //nolint:errcheck
}

// GET /lookup
func (h *Handler) LookupBySKU(w http.ResponseWriter, r *http.Request) {
	sku := r.URL.Query().Get("sku")
	if sku == "" {
		jsonError(w, "SKU is required", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.LookupBySKU(username, sku)
	if err != nil {
		switch err.Error() {
		case "item not found", "model not found for this item", "product not found for this model", "employee or location not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		default:
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /invoice/statuses
func (h *Handler) GetInvoiceStatuses(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetInvoiceStatuses()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// PUT /invoice/{invoice_id}/status
func (h *Handler) UpdateInvoiceStatus(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	var body struct {
		StatusInvoiceID int `json:"status_invoice_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.StatusInvoiceID == 0 {
		jsonError(w, "status_invoice_id is required", http.StatusBadRequest)
		return
	}
	result, err := h.svc.UpdateInvoiceStatus(invoiceID, body.StatusInvoiceID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// PUT /invoice/{invoice_id}
func (h *Handler) AddItemsToInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	var input invSvc.AddItemsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.AddItemsToInvoice(username, invoiceID, input)
	if err != nil {
		switch err.Error() {
		case "employee or location not found", "invoice not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		case "invoice is finalized (locked) and cannot be updated":
			jsonError(w, err.Error(), http.StatusForbidden)
		default:
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /invoices/search
func (h *Handler) SearchInvoices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.SearchInvoices(username, q)
	if err != nil {
		switch err.Error() {
		case "employee or location not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		default:
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// DELETE /invoice/{invoice_id}
func (h *Handler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	if err := h.svc.DeleteInvoice(username, invoiceID); err != nil {
		switch err.Error() {
		case "employee or location not found", "invoice not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		case "invoice is finalized (locked) and cannot be updated":
			jsonError(w, err.Error(), http.StatusForbidden)
		default:
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	jsonResponse(w, map[string]string{"message": "Invoice deleted successfully"}, http.StatusOK)
}

// PUT /invoice/finalize/{invoice_id}
func (h *Handler) FinalizeInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	if err := h.svc.FinalizeInvoice(invoiceID); err != nil {
		switch err.Error() {
		case "invoice not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		default:
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	jsonResponse(w, map[string]string{"message": "Invoice finalized successfully"}, http.StatusOK)
}

// PUT /invoice/unfinalize/{invoice_id}
func (h *Handler) UnfinalizeInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	if err := h.svc.UnfinalizeInvoice(invoiceID); err != nil {
		switch err.Error() {
		case "invoice not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		default:
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	jsonResponse(w, map[string]string{"message": "Invoice unfinalized successfully"}, http.StatusOK)
}

// ─── Invoice Price Book ──────────────────────────────────────────────────────

func qint(vals ...string) *int {
	for _, v := range vals {
		if v == "" {
			continue
		}
		n, err := strconv.Atoi(v)
		if err == nil {
			return &n
		}
	}
	return nil
}

// GET /invoice/{invoice_id}/price-book/lens/list
func (h *Handler) InvPBLensList(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	q := r.URL.Query()
	f := invSvc.InvPBLensFilters{}
	if v := qint(q.Get("brand_id")); v != nil {
		f.BrandID = v
	}
	if v := qint(q.Get("vendor_id")); v != nil {
		f.VendorID = v
	}
	if v := qint(q.Get("type_id")); v != nil {
		f.TypeID = v
	}
	if v := qint(q.Get("material_id")); v != nil {
		f.MaterialID = v
	}
	if v := qint(q.Get("special_feature_id")); v != nil {
		f.SpecialFeatureID = v
	}
	if v := qint(q.Get("series_id")); v != nil {
		f.SeriesID = v
	}
	if v := q.Get("search"); v != "" {
		f.Search = &v
	}
	if v := q.Get("source"); v != "" {
		f.Source = &v
	}
	if v := qint(q.Get("page")); v != nil {
		f.Page = *v
	} else {
		f.Page = 1
	}
	if v := qint(q.Get("per_page")); v != nil {
		f.PerPage = *v
	} else {
		f.PerPage = 25
	}

	result, err := h.svc.InvoicePBLenses(invoiceID, f)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /invoice/{invoice_id}/price-book/treatment/list
func (h *Handler) InvPBTreatmentList(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	var search *string
	if v := r.URL.Query().Get("search"); v != "" {
		search = &v
	}
	var empID int64
	username := pkgAuth.UsernameFromContext(r.Context())
	if username != "" {
		h.svc.DB().Raw(`
			SELECT e.id_employee FROM employee_login el
			JOIN employee e ON e.employee_login_id = el.id_employee_login
			WHERE el.employee_login = ?`, username).Scan(&empID)
	}
	result, err := h.svc.InvoicePBTreatments(invoiceID, search, empID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /invoice/{invoice_id}/price-book/additional/list
func (h *Handler) InvPBAddServiceList(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	q := r.URL.Query()
	var typeID *int
	if v := qint(q.Get("type_id")); v != nil {
		typeID = v
	}
	var search *string
	if v := q.Get("search"); v != "" {
		search = &v
	}
	result, err := h.svc.InvoicePBAddServices(invoiceID, typeID, search)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

package daily_close_handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	dailyCloseSvc "sighthub-backend/internal/services/daily_close_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *dailyCloseSvc.Service }

func New(svc *dailyCloseSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func httpStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"),
		strings.Contains(msg, "already exists"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// POST /count_sheet
func (h *Handler) CreateDailyClose(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var input dailyCloseSvc.CreateDailyCloseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.CreateDailyClose(username, input)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /count_sheet
func (h *Handler) GetDailyCloseSummary(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	dateStr := r.URL.Query().Get("date")

	result, err := h.svc.GetDailyCloseSummary(username, dateStr)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// PUT /count_sheet
func (h *Handler) UpdateDailyClose(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var input dailyCloseSvc.CreateDailyCloseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.UpdateDailyClose(username, input)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /payment_methods
func (h *Handler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetPaymentMethods()
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, map[string]interface{}{"payment_methods": result})
}

// GET /daily_close_detail
func (h *Handler) GetDailyCloseDetail(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	dateStr := r.URL.Query().Get("date")

	result, err := h.svc.GetDailyCloseDetail(username, dateStr)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /invoices_payments
func (h *Handler) GetInvoicesPayments(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	dateStr := r.URL.Query().Get("date")

	result, err := h.svc.GetInvoicesPayments(username, dateStr)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /payments_summary
func (h *Handler) GetPaymentsSummary(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	dateStr := r.URL.Query().Get("date")

	result, err := h.svc.GetPaymentsSummary(username, dateStr)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /transfer_credit_summary
func (h *Handler) GetTransferCreditSummary(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	dateStr := r.URL.Query().Get("date")

	var invoiceID, patientID *int64
	if v := r.URL.Query().Get("invoice_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			invoiceID = &id
		}
	}
	if v := r.URL.Query().Get("patient_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			patientID = &id
		}
	}

	result, err := h.svc.GetTransferCreditSummary(username, dateStr, invoiceID, patientID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /daily_close_report_html
func (h *Handler) RenderDailyCloseReportHTML(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	dateStr := r.URL.Query().Get("date")

	data, err := h.svc.RenderDailyCloseReportHTML(username, dateStr)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}

	templatesDir := os.Getenv("PDF_TEMPLATES_DIR")
	if templatesDir == "" {
		templatesDir = "internal/templates/pdf"
	}

	tmpl, err := template.ParseFiles(filepath.Join(templatesDir, "daily_close.html"))
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": "template error: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	tmpl.Execute(w, data)
}

// GET /count_sheet_html
func (h *Handler) RenderCountSheetHTML(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	dateStr := r.URL.Query().Get("date")

	data, err := h.svc.RenderCountSheetHTML(username, dateStr)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}

	templatesDir := os.Getenv("PDF_TEMPLATES_DIR")
	if templatesDir == "" {
		templatesDir = "internal/templates/pdf"
	}

	tmpl, err := template.ParseFiles(filepath.Join(templatesDir, "count_sheet.html"))
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": "template error: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	tmpl.Execute(w, data)
}

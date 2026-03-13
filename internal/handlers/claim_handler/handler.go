package claim_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	claimSvc "sighthub-backend/internal/services/claim_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *claimSvc.Service }

func New(svc *claimSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func errStatus(code int) int { return code }

func httpStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"),
		strings.Contains(msg, "cannot"), strings.Contains(msg, "must be"):
		return http.StatusBadRequest
	case strings.Contains(msg, "forbidden"), strings.Contains(msg, "finalized"):
		return http.StatusForbidden
	case strings.Contains(msg, "attached"), strings.Contains(msg, "status"):
		return http.StatusBadRequest
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

// GET /invoices
func (h *Handler) GetInsuranceInvoices(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 500 {
		perPage = 100
	}

	params := map[string]string{
		"status":               r.URL.Query().Get("status"),
		"date_start":           r.URL.Query().Get("date_start"),
		"date_end":             r.URL.Query().Get("date_end"),
		"invoice_number":       r.URL.Query().Get("invoice_number"),
		"insurance_company_id": r.URL.Query().Get("insurance_company_id"),
		"patient_name":         r.URL.Query().Get("patient_name"),
	}

	result, err := h.svc.GetInsuranceInvoices(username, params, page, perPage)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /insurance-companies
func (h *Handler) GetInsuranceCompanies(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetInsuranceCompanies()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /insurance-payment-types
func (h *Handler) GetInsurancePaymentTypes(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetInsurancePaymentTypes()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /invoices/:invoice_id/insurance-payment
func (h *Handler) GetInvoicePaymentSummary(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	result, err := h.svc.GetInvoicePaymentSummary(invoiceID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result) // result may be nil → JSON null
}

// POST /invoices/:invoice_id/insurance-payment
func (h *Handler) AddInsurancePayment(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	var input claimSvc.AddInsurancePaymentInput
	if v, ok := data["payment_type_id"].(float64); ok {
		input.PaymentTypeID = int(v)
	}
	if v, ok := data["amount"].(float64); ok {
		input.Amount = v
	}
	if v, ok := data["reference_number"].(string); ok {
		input.ReferenceNumber = v
	}
	if v, ok := data["note"].(string); ok {
		input.Note = v
	}
	if v, ok := data["adjust"].(float64); ok {
		input.Adjust = v
	}

	result, err := h.svc.AddInsurancePayment(username, invoiceID, input)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /super-bill/:invoice_id
func (h *Handler) GetSuperBill(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	result, err := h.svc.GetSuperBill(invoiceID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result) // nil → JSON null
}

// PUT /super-bill/:invoice_id
func (h *Handler) UpdateSuperBill(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	services, _ := data["services"].(map[string]interface{})
	diagnoses, _ := data["diagnoses"].(map[string]interface{})
	if services == nil {
		services = map[string]interface{}{}
	}
	if diagnoses == nil {
		diagnoses = map[string]interface{}{}
	}

	if err := h.svc.UpdateSuperBill(invoiceID, services, diagnoses); err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Include flags updated successfully"})
}

// GET /invoices/:invoice_id/patient
func (h *Handler) GetInvoicePatient(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	result, err := h.svc.GetInvoicePatient(invoiceID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /invoices/:invoice_id/responsible-party
func (h *Handler) GetResponsibleParty(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	result, err := h.svc.GetResponsibleParty(invoiceID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// PUT /invoices/:invoice_id/insurance-status
func (h *Handler) UpdateInsuranceStatus(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	status, _ := data["status"].(string)
	if status == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "status is required"})
		return
	}

	result, err := h.svc.UpdateInsuranceStatus(username, invoiceID, status)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /invoices/:invoice_id/secondary-insurance
func (h *Handler) GetSecondaryInsurance(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	result, err := h.svc.GetSecondaryInsurance(invoiceID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result) // nil → JSON null
}

// GET /invoices/:invoice_id/insurance-info
func (h *Handler) GetInvoiceInsuranceInfo(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	result, err := h.svc.GetInvoiceInsuranceInfo(invoiceID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result) // nil → JSON null
}

// GET /invoices/:invoice_id/claim-info
func (h *Handler) GetClaimInfo(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := pathInt64(r, "invoice_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	result, err := h.svc.GetClaimInfo(invoiceID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /doctors
func (h *Handler) GetDoctors(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetDoctors()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /templates
func (h *Handler) ListClaimTemplates(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.ListClaimTemplates()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /templates/:template_id
func (h *Handler) GetClaimTemplate(w http.ResponseWriter, r *http.Request) {
	templateID, err := pathInt(r, "template_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid template_id"})
		return
	}
	result, err := h.svc.GetClaimTemplate(templateID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// POST /templates
func (h *Handler) CreateClaimTemplate(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	result, err := h.svc.CreateClaimTemplate(data)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

// PUT /templates/:template_id
func (h *Handler) UpdateClaimTemplate(w http.ResponseWriter, r *http.Request) {
	templateID, err := pathInt(r, "template_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid template_id"})
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	result, err := h.svc.UpdateClaimTemplate(templateID, data)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// DELETE /templates/:template_id
func (h *Handler) DeleteClaimTemplate(w http.ResponseWriter, r *http.Request) {
	templateID, err := pathInt(r, "template_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid template_id"})
		return
	}
	if err := h.svc.DeleteClaimTemplate(templateID); err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Template deleted"})
}

// POST /templates/:template_id/render
func (h *Handler) RenderClaimPDF(w http.ResponseWriter, r *http.Request) {
	templateID, err := pathInt(r, "template_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid template_id"})
		return
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	color := "black"
	if c, ok := body["color"].(string); ok {
		color = strings.ToLower(c)
	}
	if color != "red" && color != "black" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "color must be 'red' or 'black'"})
		return
	}

	templateFields, _ := body["template_fields"].(map[string]interface{})
	otherFields, _ := body["other_fields"].(map[string]interface{})
	if templateFields == nil {
		templateFields = map[string]interface{}{}
	}
	if otherFields == nil {
		otherFields = map[string]interface{}{}
	}

	pdfBytes, filename, err := h.svc.RenderClaimPDF(templateID, templateFields, otherFields, color)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

// GET /templates/:template_id/pdf
func (h *Handler) GetTemplatePDF(w http.ResponseWriter, r *http.Request) {
	templateID, err := pathInt(r, "template_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid template_id"})
		return
	}
	pdfBytes, filename, err := h.svc.GetTemplatePDF(templateID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

// GET /templates/:template_id/preview
func (h *Handler) PreviewTemplatePDF(w http.ResponseWriter, r *http.Request) {
	templateID, err := pathInt(r, "template_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid template_id"})
		return
	}
	pdfBytes, filename, err := h.svc.PreviewTemplatePDF(templateID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

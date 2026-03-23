package ticket_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	pkgAuth "sighthub-backend/pkg/auth"
	ticketSvc "sighthub-backend/internal/services/ticket_service"
)

type Handler struct{ svc *ticketSvc.Service }

func New(svc *ticketSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func errorStatus(err error) int {
	if strings.Contains(err.Error(), "not found") {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

// POST /invoice/{invoice_id}
func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["invoice_id"]
	invoiceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req ticketSvc.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	result, err := h.svc.CreateTicket(username, invoiceID, &req)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

// PUT /{id_lab_ticket}
func (h *Handler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id_lab_ticket"]
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid id_lab_ticket"})
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		jsonResponse(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req ticketSvc.UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	result, err := h.svc.UpdateTicket(username, ticketID, &req)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// POST /{ticket_id}/notify-patient
func (h *Handler) NotifyPatient(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["ticket_id"]
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid ticket_id"})
		return
	}
	result, err := h.svc.NotifyPatient(ticketID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// ── GET endpoints ───────────────────────────────────────────────────────────

// GET /
func (h *Handler) ListTickets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	locID := q.Get("location_id")
	dateFrom := q.Get("date_from")
	dateTo := q.Get("date_to")
	statusID := q.Get("status_id")

	result, err := h.svc.ListTickets(ticketSvc.ListTicketsParams{
		LocationID: strPtr(locID),
		DateFrom:   strPtr(dateFrom),
		DateTo:     strPtr(dateTo),
		StatusID:   strPtr(statusID),
	})
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /invoice/{invoice_id}
func (h *Handler) GetTicketsByInvoice(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["invoice_id"]
	invoiceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice_id"})
		return
	}
	result, err := h.svc.GetTicketsByInvoice(invoiceID)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /{id_lab_ticket}
func (h *Handler) GetTicketByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id_lab_ticket"]
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid id_lab_ticket"})
		return
	}
	result, err := h.svc.GetTicketByID(ticketID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /search
func (h *Handler) SearchTickets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	result, err := h.svc.SearchTickets(q.Get("ticket_number"), q.Get("invoice_number"))
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /vendor_brand
func (h *Handler) GetVendorBrand(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetVendorBrandCombinations()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /products
func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	result, err := h.svc.GetProducts(q.Get("vendor_id"), q.Get("brand_id"))
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// ── Reference data handlers ─────────────────────────────────────────────────

func (h *Handler) GetStatuses(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetStatuses)
}

func (h *Handler) GetLensStatuses(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLensStatuses)
}

func (h *Handler) GetLabs(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLabs)
}

func (h *Handler) GetFrameTypeMaterials(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetFrameTypeMaterials)
}

func (h *Handler) GetFrameShapes(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetFrameShapes)
}

func (h *Handler) GetLensTypes(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLensTypes)
}

func (h *Handler) GetLensMaterials(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLensMaterials)
}

func (h *Handler) AddLensMaterial(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MaterialName string `json:"material_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	result, err := h.svc.AddLensMaterial(body.MaterialName)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ticketSvc.ErrMaterialNameRequired {
			status = http.StatusBadRequest
		}
		jsonResponse(w, status, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

func (h *Handler) GetLensStyles(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLensStyles)
}

func (h *Handler) GetLensTintColors(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLensTintColors)
}

func (h *Handler) GetLensSampleColors(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLensSampleColors)
}

func (h *Handler) GetLensSafetyThicknesses(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLensSafetyThicknesses)
}

func (h *Handler) GetLensBevels(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLensBevels)
}

func (h *Handler) GetLensEdges(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLensEdges)
}

func (h *Handler) GetLensSeries(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetLensSeries)
}

func (h *Handler) GetContactServices(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetContactServices)
}

func (h *Handler) GetContactLensBrands(w http.ResponseWriter, r *http.Request) {
	h.refData(w, h.svc.GetContactLensBrands)
}

// ── helpers ─────────────────────────────────────────────────────────────────

func (h *Handler) refData(w http.ResponseWriter, fn func() ([]map[string]interface{}, error)) {
	result, err := fn()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// GET /api/ticket/{ticket_id}/print
func (h *Handler) PrintTicket(w http.ResponseWriter, r *http.Request) {
	ticketID, _ := strconv.ParseInt(mux.Vars(r)["ticket_id"], 10, 64)
	if ticketID == 0 {
		jsonResponse(w, 400, map[string]string{"error": "invalid ticket_id"})
		return
	}
	includeInvoice := r.URL.Query().Get("include_invoice") == "true"
	result, err := h.svc.PrintTicket(ticketID, includeInvoice)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /api/ticket/{ticket_id}/lens_options
func (h *Handler) GetTicketLensOptions(w http.ResponseWriter, r *http.Request) {
	ticketID, _ := strconv.ParseInt(mux.Vars(r)["ticket_id"], 10, 64)
	if ticketID == 0 {
		jsonResponse(w, 400, map[string]string{"error": "invalid ticket_id"})
		return
	}
	result, err := h.svc.GetTicketLensOptions(ticketID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

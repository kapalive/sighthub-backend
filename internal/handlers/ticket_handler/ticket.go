package ticket_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	pkgAuth "sighthub-backend/pkg/auth"
	ticketSvc "sighthub-backend/internal/services/ticket_service"
	vwSvc "sighthub-backend/internal/services/visionweb_service"
	zeissSvc "sighthub-backend/internal/services/zeiss_service"
)

type Handler struct {
	svc      *ticketSvc.Service
	vwSvc    *vwSvc.Service
	zeissSvc *zeissSvc.CatalogService
}

func New(svc *ticketSvc.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) SetVWService(s *vwSvc.Service)            { h.vwSvc = s }
func (h *Handler) SetZeissService(s *zeissSvc.CatalogService) { h.zeissSvc = s }

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
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	if err := dec.Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
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

	// If ticket has VW order, fetch status from VisionWeb
	if h.vwSvc != nil {
		if vwID, ok := result["vw_order_id"].(*string); ok && vwID != nil && *vwID != "" {
			if status, err := h.vwSvc.GetOrderStatus(*vwID); err == nil {
				result["vw_order_status"] = status
			} else {
				result["vw_order_status"] = map[string]string{"error": err.Error()}
			}
		}
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
	username := pkgAuth.UsernameFromContext(r.Context())
	data, err := h.svc.GetLabsForEmployee(username)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, data)
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

// GET /api/ticket/{ticket_id}/order-requirements
// Unified endpoint: delegates to Zeiss or VisionWeb based on ticket lab_id.
// Returns lens_source so frontend knows whether to show Order button.
func (h *Handler) GetOrderRequirements(w http.ResponseWriter, r *http.Request) {
	ticketID, _ := strconv.ParseInt(mux.Vars(r)["ticket_id"], 10, 64)
	if ticketID == 0 {
		jsonResponse(w, 400, map[string]string{"error": "invalid ticket_id"})
		return
	}

	lensSource := h.svc.GetTicketLensSource(ticketID)

	// Custom or no lens — no electronic order possible
	if lensSource == "" || lensSource == "custom" {
		jsonResponse(w, 200, map[string]interface{}{
			"ready":       false,
			"lens_source": lensSource,
			"can_order":   false,
			"message":     "Manual order — electronic submission not available for custom lenses",
		})
		return
	}

	// Check lab_id to determine vendor
	labID, err := h.svc.GetTicketLabID(ticketID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}

	if labID != nil && *labID == zeissSvc.ZeissVendorID {
		// ── Zeiss path ──
		if h.zeissSvc == nil {
			jsonResponse(w, 500, map[string]string{"error": "Zeiss service not configured"})
			return
		}
		username := pkgAuth.UsernameFromContext(r.Context())
		empID, err := h.svc.EmployeeIDByUsername(username)
		if err != nil {
			jsonResponse(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		result := h.zeissSvc.CheckZeissOrderRequirements(ticketID, empID)
		jsonResponse(w, 200, map[string]interface{}{
			"ready":       result.Ready,
			"lens_source": lensSource,
			"can_order":   true,
			"provider":    "zeiss",
			"fields":      result.Fields,
		})
		return
	}

	// ── VisionWeb path ──
	if h.vwSvc == nil {
		jsonResponse(w, 500, map[string]string{"error": "VisionWeb service not configured"})
		return
	}
	result := h.vwSvc.CheckOrderRequirements(ticketID)
	jsonResponse(w, 200, map[string]interface{}{
		"ready":       result.Ready,
		"lens_source": lensSource,
		"can_order":   true,
		"provider":    "vision_web",
		"fields":      result.Fields,
	})
}

// POST /api/ticket/{ticket_id}/order
func (h *Handler) PlaceVWOrder(w http.ResponseWriter, r *http.Request) {
	ticketID, _ := strconv.ParseInt(mux.Vars(r)["ticket_id"], 10, 64)
	if ticketID == 0 {
		jsonResponse(w, 400, map[string]string{"error": "invalid ticket_id"})
		return
	}

	// Check lens source — custom lenses cannot be ordered electronically
	lensSource := h.svc.GetTicketLensSource(ticketID)
	if lensSource == "" || lensSource == "custom" {
		jsonResponse(w, 400, map[string]string{"error": "manual orders cannot be submitted electronically"})
		return
	}

	// Check lab_id — route to correct provider
	labID, err := h.svc.GetTicketLabID(ticketID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	if labID != nil && *labID == zeissSvc.ZeissVendorID {
		jsonResponse(w, 501, map[string]string{"error": "Zeiss order submission not yet implemented"})
		return
	}

	if h.vwSvc == nil {
		jsonResponse(w, 500, map[string]string{"error": "VisionWeb service not configured"})
		return
	}

	result, err := h.vwSvc.PlaceOrder(ticketID)
	if err != nil {
		if ve, ok := err.(*vwSvc.ValidationErrors); ok {
			jsonResponse(w, 422, map[string]interface{}{
				"error":  "validation failed",
				"errors": ve.Errors,
			})
			return
		}
		// Already submitted — return 409 with order info
		if result != nil && result.VWOrderID != "" {
			jsonResponse(w, 409, map[string]interface{}{
				"error":       err.Error(),
				"vw_order_id": result.VWOrderID,
				"status":      result.Status,
			})
			return
		}
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

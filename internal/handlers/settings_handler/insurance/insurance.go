package insurance

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	svc "sighthub-backend/internal/services/settings_service"
)

type Handler struct{ svc *svc.Service }

func New(s *svc.Service) *Handler { return &Handler{svc: s} }

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func pathID(r *http.Request) int {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	return id
}

// ── Insurance Companies ────────────────────────────────────────────────────

func (h *Handler) ListCompanies(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListInsuranceCompanies()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CompanyName string `json:"company_name"`
		CoverageID  *int   `json:"insurance_coverage_type_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if body.CompanyName == "" {
		jsonError(w, "company_name is required", 400)
		return
	}
	data, err := h.svc.CreateInsuranceCompany(body.CompanyName, body.CoverageID)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.UpdateInsuranceCompany(pathID(r), body)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteInsuranceCompany(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Coverage Types ─────────────────────────────────────────────────────────

func (h *Handler) ListCoverageTypes(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListCoverageTypes()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ── Insurance Types ────────────────────────────────────────────────────────

func (h *Handler) ListInsuranceTypes(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListInsuranceTypes()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ── Insurance Payment Types ────────────────────────────────────────────────

func (h *Handler) ListPaymentTypes(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListInsurancePaymentTypes()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreatePaymentType(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateInsurancePaymentType(body.Name, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdatePaymentType(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateInsurancePaymentType(pathID(r), body); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeletePaymentType(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteInsurancePaymentType(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

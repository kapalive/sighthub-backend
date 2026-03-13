package questionnaire_handler

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	svc "sighthub-backend/internal/services/questionnaire_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *svc.Service
	db  *gorm.DB
}

func New(s *svc.Service, db *gorm.DB) *Handler { return &Handler{svc: s, db: db} }

func jsonOK(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// POST /api/questionnaire/referral
func (h *Handler) CreateReferral(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	emp, locID, err := h.svc.GetEmployeeAndLocation(username)
	if err != nil || emp == nil || locID == 0 {
		jsonError(w, "Employee or Store not found", http.StatusNotFound)
		return
	}

	var input svc.CreateReferralInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	input.LocationID = locID
	input.EmployeeID = emp.IDEmployee

	rec, err := h.svc.CreateReferral(input)
	if err != nil {
		if err.Error() == "missing required fields" {
			jsonError(w, "Missing required fields", http.StatusBadRequest)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]interface{}{
		"message": "Questionnaire referral created successfully",
		"record":  rec,
	})
}

// GET /api/questionnaire/referral_sources
func (h *Handler) GetReferralSources(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetReferralSources()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, list)
}

// GET /api/questionnaire/visit_reasons
func (h *Handler) GetVisitReasons(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetVisitReasons()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, list)
}

package rx

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	rxSvc "sighthub-backend/internal/services/patient_service/rx"
	pkgAuth "sighthub-backend/pkg/auth"
)

// ─── Handler ──────────────────────────────────────────────────────────────────

type Handler struct {
	svc *rxSvc.Service
}

func New(svc *rxSvc.Service) *Handler { return &Handler{svc: svc} }

// ─── Utils ────────────────────────────────────────────────────────────────────

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

// ─── GET /rx/amidoctor ────────────────────────────────────────────────────────

func (h *Handler) AmIDoctor(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	result, err := h.svc.AmIDoctor(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result == nil {
		jsonResponse(w, map[string]interface{}{"doctor": nil, "npi": nil, "ein": nil}, http.StatusOK)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GET /latest-rx ───────────────────────────────────────────────────────────

func (h *Handler) GetLatestRx(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id_patient")
	if idStr == "" {
		jsonResponse(w, map[string]interface{}{}, http.StatusOK)
		return
	}
	patientID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonResponse(w, map[string]interface{}{}, http.StatusOK)
		return
	}

	result, err := h.svc.GetLatestRx(patientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GET /rx-list ─────────────────────────────────────────────────────────────

func (h *Handler) GetRxList(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.svc.GetRxList(patientID)
	if err != nil {
		if err.Error() == "patient not found" {
			jsonError(w, err.Error(), http.StatusNotFound)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GET /rx ──────────────────────────────────────────────────────────────────

func (h *Handler) GetRx(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id_rx")
	if idStr == "" {
		jsonError(w, "id_rx is required", http.StatusBadRequest)
		return
	}
	rxID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid id_rx", http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetRx(rxID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── POST /rx ─────────────────────────────────────────────────────────────────

func (h *Handler) CreateRx(w http.ResponseWriter, r *http.Request) {
	var input rxSvc.CreateRxInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if input.PatientID == 0 {
		jsonError(w, "patient_id required", http.StatusBadRequest)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.CreateRx(username, input)
	if err != nil {
		switch err.Error() {
		case "employee or location not found", "patient not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		default:
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// ─── PUT /rx ──────────────────────────────────────────────────────────────────

func (h *Handler) UpdateRx(w http.ResponseWriter, r *http.Request) {
	var input rxSvc.UpdateRxInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	if err := h.svc.UpdateRx(username, input); err != nil {
		switch err.Error() {
		case "prescription not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		case "employee or location not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		default:
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	jsonResponse(w, map[string]string{
		"status":  "success",
		"message": "Prescription updated successfully",
	}, http.StatusOK)
}

// ─── DELETE /rx/{id_rx} ───────────────────────────────────────────────────────

func (h *Handler) DeleteRx(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id_rx"]
	rxID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid id_rx", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteRx(rxID); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, map[string]string{
		"status":  "success",
		"message": "Prescription deleted successfully.",
	}, http.StatusOK)
}

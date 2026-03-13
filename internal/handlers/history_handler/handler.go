package history_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	histSvc "sighthub-backend/internal/services/history_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *histSvc.Service }

func New(svc *histSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func parsePathInt64(r *http.Request, key string) (int64, bool) {
	v, ok := mux.Vars(r)[key]
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	return n, err == nil
}

// POST /{exam_id}
func (h *Handler) SaveHistory(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input histSvc.SaveHistoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "missing data"})
		return
	}

	result, err := h.svc.SaveHistory(username, examID, input)
	if err != nil {
		switch err.Error() {
		case "employee not found", "exam not found", "patient not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "you are not authorized to update this exam", "cannot create history for a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message":        "History saved successfully",
		"id_history_eye": result["id_history_eye"],
		"data":           result,
	})
}

// GET /{exam_id}
func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	result, err := h.svc.GetHistory(examID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// PUT /{exam_id}
func (h *Handler) UpdateHistory(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input histSvc.UpdateHistoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.UpdateHistory(username, examID, input)
	if err != nil {
		switch err.Error() {
		case "employee not found", "exam not found", "patient not found",
			"history record not found for this exam":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "cannot update history for a completed exam", "you are not authorized to update this exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case "medical_record is required", "failed to create medical record":
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{"message": "History updated successfully", "data": result})
}

// GET /races
func (h *Handler) GetRaces(w http.ResponseWriter, r *http.Request) {
	races, err := h.svc.GetRaces()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{"races": races})
}

// GET /ethnicities
func (h *Handler) GetEthnicities(w http.ResponseWriter, r *http.Request) {
	ethnicities, err := h.svc.GetEthnicities()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{"ethnicities": ethnicities})
}

// GET /{exam_id}/patient-info
func (h *Handler) GetPatientInfo(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	result, err := h.svc.GetPatientInfo(examID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// POST /medications/{exam_id}
func (h *Handler) SaveMedication(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "field 'title' is required"})
		return
	}

	if err := h.svc.SaveMedication(username, examID, body.Title); err != nil {
		switch err.Error() {
		case "employee not found", "exam not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "cannot modify a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]string{"message": "Medication saved successfully"})
}

// GET /medications/{exam_id}
func (h *Handler) GetMedications(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	meds, err := h.svc.GetMedications(username, examID)
	if err != nil {
		switch err.Error() {
		case "employee not found", "exam not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, meds)
}

// DELETE /medications/{exam_id}/{medication_id}
func (h *Handler) DeleteMedication(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok1 := parsePathInt64(r, "exam_id")
	medicationID, ok2 := parsePathInt64(r, "medication_id")
	if !ok1 || !ok2 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid path parameters"})
		return
	}

	if err := h.svc.DeleteMedication(username, examID, medicationID); err != nil {
		switch err.Error() {
		case "employee not found", "exam not found", "medication not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "cannot modify a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Medication deleted successfully"})
}

// POST /allergies/{exam_id}
func (h *Handler) SaveAllergy(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "allergy 'title' is required"})
		return
	}

	if err := h.svc.SaveAllergy(username, examID, body.Title); err != nil {
		switch err.Error() {
		case "employee not found", "exam not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "cannot modify a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]string{"message": "Allergy saved successfully"})
}

// GET /allergies/{exam_id}
func (h *Handler) GetAllergies(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	allergies, err := h.svc.GetAllergies(username, examID)
	if err != nil {
		switch err.Error() {
		case "employee not found", "exam not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, allergies)
}

// DELETE /allergies/{exam_id}/{allergy_id}
func (h *Handler) DeleteAllergy(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok1 := parsePathInt64(r, "exam_id")
	allergyID, ok2 := parsePathInt64(r, "allergy_id")
	if !ok1 || !ok2 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid path parameters"})
		return
	}

	if err := h.svc.DeleteAllergy(username, examID, allergyID); err != nil {
		switch err.Error() {
		case "employee not found", "exam not found", "allergy not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "cannot modify a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Allergy deleted successfully"})
}

package preliminary_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	prelimSvc "sighthub-backend/internal/services/preliminary_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *prelimSvc.Service }

func New(svc *prelimSvc.Service) *Handler { return &Handler{svc: svc} }

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
func (h *Handler) SavePreliminary(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input prelimSvc.SavePreliminaryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.SavePreliminary(username, examID, input)
	if err != nil {
		switch err.Error() {
		case "employee not found", "exam not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "you are not authorized to update this exam",
			"cannot create preliminary for a completed exam",
			"preliminary already exists for this exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message":                 result["message"],
		"id_preliminary_eye_exam": result["id_preliminary_eye_exam"],
		"data":                    result["data"],
	})
}

// GET /{exam_id}
func (h *Handler) GetPreliminary(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	result, err := h.svc.GetPreliminary(examID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// PUT /{exam_id}
func (h *Handler) UpdatePreliminary(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input prelimSvc.UpdatePreliminaryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.UpdatePreliminary(username, examID, input)
	if err != nil {
		switch err.Error() {
		case "employee not found", "exam not found", "preliminary record not found for this exam":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "you are not authorized to update this exam",
			"cannot update preliminary for a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": result["message"],
		"data":    result["data"],
	})
}

// GET /prescription_list?id_patient=X
func (h *Handler) GetPrescriptionList(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id_patient")
	if idStr == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "id_patient is required"})
		return
	}
	patientID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid id_patient"})
		return
	}

	result, err := h.svc.GetPrescriptionList(patientID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{"data": result})
}

// POST /{exam_id}/entrance_rx
func (h *Handler) FillEntranceRx(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input prelimSvc.FillEntranceRxInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.FillEntranceRx(examID, input)
	if err != nil {
		switch err.Error() {
		case "exam not found", "glasses prescription not found", "contact lens prescription not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "cannot modify entrance rx for a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

// PUT /{exam_id}/entrance_rx
func (h *Handler) UpdateEntranceRx(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input prelimSvc.FillEntranceRxInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.UpdateEntranceRx(examID, input)
	if err != nil {
		switch err.Error() {
		case "exam not found", "preliminary record not found for this exam",
			"glasses prescription not found", "contact lens prescription not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "cannot modify entrance rx for a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /{exam_id}/entrance_rx
func (h *Handler) GetEntranceRx(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	result, err := h.svc.GetEntranceRx(examID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// DELETE /{exam_id}/entrance_rx
func (h *Handler) DeleteEntranceRx(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input prelimSvc.FillEntranceRxInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.DeleteEntranceRx(examID, input)
	if err != nil {
		switch err.Error() {
		case "exam not found", "preliminary record not found for this exam":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "cannot modify entrance rx for a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// POST /{exam_id}/near_point_testing
func (h *Handler) CreateNearPointTesting(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input prelimSvc.NearPointTestingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.CreateNearPointTesting(examID, input)
	if err != nil {
		switch err.Error() {
		case "exam not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

// PUT /{exam_id}/near_point_testing
func (h *Handler) UpdateNearPointTesting(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input prelimSvc.NearPointTestingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.UpdateNearPointTesting(username, examID, input)
	if err != nil {
		switch err.Error() {
		case "employee not found", "exam not found",
			"preliminary record not found for this exam",
			"near point testing record not found for this exam",
			"near point testing record not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "you are not authorized to update this exam",
			"cannot update near point testing for a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /{exam_id}/near_point_testing
func (h *Handler) GetNearPointTesting(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	result, err := h.svc.GetNearPointTesting(examID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

package assessment_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	assessmentSvc "sighthub-backend/internal/services/assessment_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *assessmentSvc.Service }

func New(svc *assessmentSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func parsePathInt64(r *http.Request, key string) (int64, error) {
	v, ok := mux.Vars(r)[key]
	if !ok {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(v, 10, 64)
}

func errorStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "employee not found"),
		strings.Contains(msg, "exam not found"),
		strings.Contains(msg, "assessment not found"),
		strings.Contains(msg, "diagnosis not found"),
		strings.Contains(msg, "pqrs not found"),
		strings.Contains(msg, "disease not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "not authorized"),
		strings.Contains(msg, "cannot update"):
		return http.StatusForbidden
	case strings.Contains(msg, "assessments required"),
		strings.Contains(msg, "already in top diseases"),
		strings.Contains(msg, "assessment does not belong"),
		strings.Contains(msg, "or not authorized"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handler) SaveAssessment(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input assessmentSvc.SaveAssessmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	results, err := h.svc.SaveAssessment(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message":     "Assessments saved successfully",
		"assessments": results,
	})
}

func (h *Handler) GetAssessments(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	results, err := h.svc.GetAssessments(examID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"id_eye_exam": examID,
		"assessments": results,
	})
}

func (h *Handler) UpdateAssessment(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input assessmentSvc.UpdateAssessmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	results, err := h.svc.UpdateAssessment(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "Assessments updated successfully",
		"assessments": results,
	})
}

func (h *Handler) DeleteAssessmentDiagnosis(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	assessmentID, err := parsePathInt64(r, "assessment_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid assessment_id"})
		return
	}
	diagnosisID, err := parsePathInt64(r, "diagnosis_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid diagnosis_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	if err := h.svc.DeleteAssessmentDiagnosis(username, examID, assessmentID, diagnosisID); err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Diagnosis deleted successfully"})
}

func (h *Handler) DeleteAssessmentPQRS(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	assessmentID, err := parsePathInt64(r, "assessment_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid assessment_id"})
		return
	}
	pqrsID, err := parsePathInt64(r, "pqrs_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid pqrs_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	if err := h.svc.DeleteAssessmentPQRS(username, examID, assessmentID, pqrsID); err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "PQRS deleted successfully"})
}

func (h *Handler) DeleteAssessment(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	assessmentID, err := parsePathInt64(r, "assessment_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid assessment_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	if err := h.svc.DeleteAssessment(username, examID, assessmentID); err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Assessment deleted successfully"})
}

func (h *Handler) SearchDiagnosis(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	if term == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Search term is required"})
		return
	}
	results := h.svc.SearchDiagnosis(term)
	jsonResponse(w, http.StatusOK, results)
}

func (h *Handler) GetMyTopDiseases(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, h.svc.GetMyTopDiseases())
}

func (h *Handler) AddMyTopDisease(w http.ResponseWriter, r *http.Request) {
	var input assessmentSvc.AddTopDiseaseInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if input.LevelID == 0 || input.Type == "" || input.Code == "" || input.Title == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Missing required fields: level_id, type, code, title"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	record, err := h.svc.AddMyTopDisease(username, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Disease added successfully",
		"data":    record,
	})
}

func (h *Handler) DeleteMyTopDisease(w http.ResponseWriter, r *http.Request) {
	diseaseID, err := parsePathInt64(r, "disease_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid disease_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	if err := h.svc.DeleteMyTopDisease(username, diseaseID); err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Disease removed from top successfully"})
}

func (h *Handler) GetAllPQRS(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, h.svc.GetAllPQRS())
}

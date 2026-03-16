package referral_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	referralSvc "sighthub-backend/internal/services/referral_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *referralSvc.Service }

func New(svc *referralSvc.Service) *Handler { return &Handler{svc: svc} }

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
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "not authorized"),
		strings.Contains(msg, "cannot update"),
		strings.Contains(msg, "cannot create"):
		return http.StatusForbidden
	case strings.Contains(msg, "invalid"),
		strings.Contains(msg, "no data provided"),
		strings.Contains(msg, "no selected parameters"),
		strings.Contains(msg, "invalid doctor_type"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// POST /{exam_id}/letter
func (h *Handler) SaveReferral(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input referralSvc.SaveReferralInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.SaveReferral(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

// GET /{exam_id}
func (h *Handler) GetReferral(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	result, err := h.svc.GetReferral(examID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /{exam_id}/letter/{letter_id}
func (h *Handler) GetReferralLetterByID(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	letterID, err := parsePathInt64(r, "letter_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid letter_id"})
		return
	}
	result, err := h.svc.GetReferralLetterByID(examID, letterID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// PUT /{exam_id}/letter/{letter_id}
func (h *Handler) UpdateReferralLetter(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	letterID, err := parsePathInt64(r, "letter_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid letter_id"})
		return
	}
	var rawData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if len(rawData) == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "no data provided"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.UpdateReferralLetter(username, examID, letterID, rawData)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// DELETE /{exam_id}/letter/{letter_id}
func (h *Handler) DeleteReferralLetterOrDoctor(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	letterID, err := parsePathInt64(r, "letter_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid letter_id"})
		return
	}
	doctorType := r.URL.Query().Get("doctor_type")
	username := pkgAuth.UsernameFromContext(r.Context())
	msg, err := h.svc.DeleteReferralLetterOrDoctor(username, examID, letterID, doctorType)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":   msg,
		"letter_id": letterID,
	})
}

// GET /doctors
func (h *Handler) GetAllReferralDoctors(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetAllReferralDoctors()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// POST /doctors
func (h *Handler) CreateReferralDoctor(w http.ResponseWriter, r *http.Request) {
	var input referralSvc.CreateDoctorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	doc, err := h.svc.CreateReferralDoctor(input)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, doc)
}

// PUT /doctors/{doctor_id}
func (h *Handler) UpdateReferralDoctor(w http.ResponseWriter, r *http.Request) {
	doctorID, err := parsePathInt64(r, "doctor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid doctor_id"})
		return
	}
	var input referralSvc.CreateDoctorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	doc, err := h.svc.UpdateReferralDoctor(doctorID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, doc)
}

// DELETE /doctors/{doctor_id}
func (h *Handler) DeleteReferralDoctor(w http.ResponseWriter, r *http.Request) {
	doctorID, err := parsePathInt64(r, "doctor_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid doctor_id"})
		return
	}
	if err := h.svc.DeleteReferralDoctor(doctorID); err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Referral doctor deleted successfully"})
}

// GET /referral-letters/{letter_id}/html
func (h *Handler) PrintReferralLetter(w http.ResponseWriter, r *http.Request) {
	letterID, err := parsePathInt64(r, "letter_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid letter_id"})
		return
	}
	html, err := h.svc.GetReferralLetterHTML(letterID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// POST /referral-letters/{letter_id}/fax
func (h *Handler) FaxReferralLetter(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotImplemented, map[string]string{"error": "FAX sending not implemented"})
}

// GET /tests-build/{exam_id}
func (h *Handler) BuildTests(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	params := map[string]string{}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	text, drawing, err := h.svc.BuildTests(examID, params)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"response": text,
		"drawing":  drawing,
	})
}

// EmailReferralLetter — POST /referral-letters/{letter_id}/email
func (h *Handler) EmailReferralLetter(w http.ResponseWriter, r *http.Request) {
	letterID, err := parsePathInt64(r, "letter_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid letter_id"})
		return
	}

	var body struct {
		ToEmail string `json:"to_email"`
		Subject string `json:"subject"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ToEmail == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "to_email is required"})
		return
	}

	if err := h.svc.SendReferralEmail(letterID, body.ToEmail, body.Subject); err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "email sent successfully", "to": body.ToEmail})
}

package super_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	superSvc "sighthub-backend/internal/services/super_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *superSvc.Service }

func New(svc *superSvc.Service) *Handler { return &Handler{svc: svc} }

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
		strings.Contains(msg, "cannot create"),
		strings.Contains(msg, "finalized"):
		return http.StatusForbidden
	case strings.Contains(msg, "already exists"),
		strings.Contains(msg, "already been completed"),
		strings.Contains(msg, "required"),
		strings.Contains(msg, "cannot delete"),
		strings.Contains(msg, "does not match"),
		strings.Contains(msg, "invalid"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// POST /{exam_id}
func (h *Handler) CreateSuperEyeExam(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input superSvc.CreateSuperInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if len(input.CptHcpcsCode) == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "'cpt_hcpcs_code' object is required"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.CreateSuperEyeExam(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

// POST /{exam_id}/invoice
func (h *Handler) CreateSuperInvoice(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input superSvc.InvoiceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if len(input.CptHcpcsCode) == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "'cpt_hcpcs_code' object is required"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.CreateSuperInvoice(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

// PUT /{exam_id}
func (h *Handler) UpdateSuperEyeExam(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input superSvc.InvoiceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if len(input.CptHcpcsCode) == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "'cpt_hcpcs_code' object is required"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.UpdateSuperEyeExam(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /{exam_id}
func (h *Handler) GetSuperEyeExam(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	result, err := h.svc.GetSuperEyeExam(examID)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// DELETE /{exam_id}/item/{item_id}/diagnosis?disease_super_bill_id=123
func (h *Handler) DeleteDiagnosisByID(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	itemID, err := parsePathInt64(r, "item_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid item_id"})
		return
	}
	dsbIDStr := r.URL.Query().Get("disease_super_bill_id")
	if dsbIDStr == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "query parameter 'disease_super_bill_id' is required"})
		return
	}
	dsbID, err := strconv.ParseInt(dsbIDStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid disease_super_bill_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.DeleteDiagnosisByID(username, examID, itemID, dsbID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// DELETE /{exam_id}/item/{item_id}
func (h *Handler) DeleteItemFromInvoice(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	itemID, err := parsePathInt64(r, "item_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid item_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.DeleteItemFromInvoice(username, examID, itemID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /prof-serv?type_id=1
func (h *Handler) GetProfessionalServices(w http.ResponseWriter, r *http.Request) {
	var typeID *int
	if v := r.URL.Query().Get("type_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid type_id"})
			return
		}
		typeID = &id
	}
	result, err := h.svc.GetProfessionalServices(typeID)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /super-bill-diseases?exam_id=123
func (h *Handler) GetSuperBillDiseases(w http.ResponseWriter, r *http.Request) {
	examIDStr := r.URL.Query().Get("exam_id")
	if examIDStr == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "exam_id is required"})
		return
	}
	examID, err := strconv.ParseInt(examIDStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	result, err := h.svc.GetSuperBillDiseases(examID)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

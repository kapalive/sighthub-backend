package recall

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	recallSvc "sighthub-backend/internal/services/patient_service/recall"
	pkgAuth "sighthub-backend/pkg/auth"
)

type ListHandler struct {
	svc *recallSvc.Service
}

func NewListHandler(svc *recallSvc.Service) *ListHandler {
	return &ListHandler{svc: svc}
}

// GET /api/patient/recall-list
func (h *ListHandler) GetRecallList(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	q := r.URL.Query()

	params := recallSvc.RecallListParams{
		Page:    intParam(q.Get("page"), 1),
		PerPage: intParam(q.Get("per_page"), 25),
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
	}

	if v := q.Get("date_from"); v != "" {
		params.DateFrom = &v
	}
	if v := q.Get("date_to"); v != "" {
		params.DateTo = &v
	}
	if v := q.Get("call_status"); v != "" {
		params.CallStatus = &v
	}
	if v := q.Get("reason"); v != "" {
		params.Reason = &v
	}
	if v := q.Get("first_name"); v != "" {
		params.FirstName = &v
	}
	if v := q.Get("last_name"); v != "" {
		params.LastName = &v
	}
	if v := q.Get("dob"); v != "" {
		params.DOB = &v
	}
	if v := q.Get("city"); v != "" {
		params.City = &v
	}
	if v := q.Get("state"); v != "" {
		params.State = &v
	}
	if v := q.Get("phone"); v != "" {
		params.Phone = &v
	}
	if v := q.Get("email"); v != "" {
		params.Email = &v
	}
	if v := q.Get("has_phone"); v != "" {
		b := v == "true" || v == "1"
		params.HasPhone = &b
	}
	if v := q.Get("preferred_language_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.PreferredLanguageID = &n
		}
	}
	if v := q.Get("insurance_company_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.InsuranceCompanyID = &n
		}
	}

	result, err := h.svc.GetRecallList(username, params)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, result)
}

// POST /api/patient/recall/{recall_id}/result
func (h *ListHandler) LogCallResult(w http.ResponseWriter, r *http.Request) {
	recallID, err := strconv.ParseInt(mux.Vars(r)["recall_id"], 10, 64)
	if err != nil {
		jsonError(w, "invalid recall_id", http.StatusBadRequest)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())

	var input recallSvc.LogCallResultInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.LogCallResult(username, recallID, input); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "recall not found" {
			code = http.StatusNotFound
		}
		jsonError(w, err.Error(), code)
		return
	}

	jsonOK(w, map[string]string{"message": "Call result logged"})
}

func intParam(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return def
	}
	return n
}

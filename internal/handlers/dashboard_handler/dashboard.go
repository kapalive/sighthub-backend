package dashboard_handler

import (
	"encoding/json"
	"net/http"

	dashSvc "sighthub-backend/internal/services/dashboard_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *dashSvc.Service }

func New(svc *dashSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// GET /weekly_income
func (h *Handler) GetWeeklyIncome(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	data, err := h.svc.GetWeeklyIncome(username)
	if err != nil {
		jsonResponse(w, 404, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, map[string]interface{}{"data": data})
}

// GET /appointments
func (h *Handler) GetAppointmentStatuses(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	data, err := h.svc.GetAppointmentStatuses(username)
	if err != nil {
		jsonResponse(w, 404, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, map[string]interface{}{"appointments": map[string]interface{}{"data": data}})
}

// GET /employee_sales
func (h *Handler) GetEmployeeSales(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "w"
	}

	data, err := h.svc.GetEmployeeSales(username, period)
	if err != nil {
		jsonResponse(w, 404, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, map[string]interface{}{"data": data})
}

// GET /employee_invoices
func (h *Handler) GetEmployeeInvoices(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "w"
	}

	data, err := h.svc.GetEmployeeInvoices(username, period)
	if err != nil {
		jsonResponse(w, 404, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, map[string]interface{}{"data": data})
}

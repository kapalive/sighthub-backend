package ar_report_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	arSvc "sighthub-backend/internal/services/ar_report_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *arSvc.Service
}

func New(svc *arSvc.Service) *Handler {
	return &Handler{svc: svc}
}

// ─── GetBalanceDue  GET /api/ar/balance_due ───────────────────────────────────

func (h *Handler) GetBalanceDue(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	locID, err := parseOptionalLocID(r, username, h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	period := 0
	if v := r.URL.Query().Get("period"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			period = p
		}
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	result, err := h.svc.GetBalanceDue(username, locID, period, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GetCredits  GET /api/ar/credits ─────────────────────────────────────────

func (h *Handler) GetCredits(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	locID, err := parseOptionalLocID(r, username, h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	result, err := h.svc.GetCredits(username, locID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GetCountSheets  GET /api/ar/count_sheets ────────────────────────────────

func (h *Handler) GetCountSheets(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var dateFrom, dateTo *time.Time
	if v := r.URL.Query().Get("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			dateFrom = &t
		} else {
			http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			dateTo = &t
		} else {
			http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	result, err := h.svc.GetCountSheets(username, dateFrom, dateTo)
	if err != nil {
		if err.Error() == "employee or location not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── CreateCountSheet  POST /api/ar/count_sheets ─────────────────────────────

func (h *Handler) CreateCountSheet(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var body struct {
		Notes string `json:"notes"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	result, err := h.svc.CreateCountSheet(username, body.Notes)
	if err != nil {
		if ae, ok := err.(*arSvc.ActiveCountSheetError); ok {
			jsonResponse(w, map[string]interface{}{
				"error":          err.Error(),
				"id_count_sheet": ae.ID,
			}, http.StatusBadRequest)
			return
		}
		switch err.Error() {
		case "employee or location not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		case "no open AR invoices found for this location":
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// ─── DeleteCountSheet  DELETE /api/ar/count_sheets ───────────────────────────

func (h *Handler) DeleteCountSheet(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	idStr := r.URL.Query().Get("id_count_sheet")
	if idStr == "" {
		http.Error(w, "id_count_sheet is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id_count_sheet", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteCountSheet(username, id); err != nil {
		if err.Error() == "employee or location not found" ||
			err.Error() == "count sheet not found or does not belong to your location" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"status":  200,
		"message": "AR count sheet " + idStr + " deleted successfully",
	}, http.StatusOK)
}

// ─── GetCountSheetInfo  GET /api/ar/count_sheets/{id} ────────────────────────

func (h *Handler) GetCountSheetInfo(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	id, err := parseCountSheetID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetCountSheetInfo(username, id)
	if err != nil {
		if err.Error() == "employee or location not found" ||
			err.Error() == "count sheet not found or does not belong to your location" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── UpdateCountSheetNotes  PUT /api/ar/count_sheets/{id} ────────────────────

func (h *Handler) UpdateCountSheetNotes(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	id, err := parseCountSheetID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body struct {
		Notes *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Notes == nil {
		http.Error(w, `field "notes" is required`, http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateCountSheetNotes(username, id, *body.Notes); err != nil {
		if err.Error() == "employee or location not found" ||
			err.Error() == "count sheet not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "Notes updated successfully"}, http.StatusOK)
}

// ─── CloseCountSheet  POST /api/ar/count_sheets/{id}/close ───────────────────

func (h *Handler) CloseCountSheet(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	id, err := parseCountSheetID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.svc.CloseCountSheet(username, id); err != nil {
		switch err.Error() {
		case "employee or location not found",
			"count sheet not found or does not belong to your location":
			http.Error(w, err.Error(), http.StatusNotFound)
		case "count sheet is already closed":
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, map[string]string{
		"message": "AR count sheet closed successfully",
	}, http.StatusOK)
}

// ─── GetCountSheetItems  GET /api/ar/count_sheets/items ──────────────────────

func (h *Handler) GetCountSheetItems(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	idStr := r.URL.Query().Get("id_count_sheet")
	if idStr == "" {
		http.Error(w, "id_count_sheet is required", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id_count_sheet", http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetCountSheetItems(username, id)
	if err != nil {
		if err.Error() == "employee or location not found" ||
			err.Error() == "count sheet not found or does not belong to your location" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── AddInvoiceToCountSheet  POST /api/ar/count_sheets/items ─────────────────

func (h *Handler) AddInvoiceToCountSheet(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var body struct {
		IDCountSheet  *int    `json:"id_count_sheet"`
		InvoiceNumber *string `json:"invoice_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "request body is required", http.StatusBadRequest)
		return
	}
	if body.IDCountSheet == nil || body.InvoiceNumber == nil {
		http.Error(w, "id_count_sheet and invoice_number are required", http.StatusBadRequest)
		return
	}

	result, err := h.svc.AddInvoiceToCountSheet(username, *body.IDCountSheet, *body.InvoiceNumber)
	if err != nil {
		switch err.Error() {
		case "employee or location not found",
			"count sheet not found or does not belong to your location",
			"invoice not found in this location",
			"invoice not found in the outstanding list for this count sheet":
			http.Error(w, err.Error(), http.StatusNotFound)
		case "count sheet is closed",
			"this invoice has already been counted in this sheet":
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// parseOptionalLocID reads location_id from query, or falls back to employee's location.
func parseOptionalLocID(r *http.Request, username string, h *Handler) (int, error) {
	if v := r.URL.Query().Get("location_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			return id, nil
		}
	}
	_, loc, err := h.svc.GetEmployeeLocation(username)
	if err != nil {
		return 0, err
	}
	return loc, nil
}

func parseCountSheetID(r *http.Request) (int, error) {
	return strconv.Atoi(mux.Vars(r)["id"])
}

func jsonResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

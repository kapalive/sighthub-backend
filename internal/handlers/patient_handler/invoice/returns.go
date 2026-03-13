package invoice

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	invSvc "sighthub-backend/internal/services/patient_service/invoice"
	pkgAuth "sighthub-backend/pkg/auth"
)

// ─── ProcessReturn  POST /invoice/{invoice_id}/return ───────────────────────

func (h *Handler) ProcessReturn(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.ProcessReturnInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.ProcessReturn(username, invoiceID, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found":
			statusCode = http.StatusNotFound
		case "invoice was created in another location. returns can only be processed in the same location":
			statusCode = http.StatusForbidden
		case "no items provided for return":
			statusCode = http.StatusBadRequest
		}
		if statusCode == http.StatusInternalServerError {
			statusCode = http.StatusConflict
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── DenyReturn  PUT /return/{return_id}/deny ─────────────────────────────────

func (h *Handler) DenyReturn(w http.ResponseWriter, r *http.Request) {
	returnID, err := parseReturnID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	result, err := h.svc.DenyReturn(username, returnID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "return record not found", "invoice not found":
			statusCode = http.StatusNotFound
		case "invoice was created in another location. return denial can only be processed in the same location":
			statusCode = http.StatusForbidden
		case "return request is already processed or in invalid status":
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── ConfirmReturn  PUT /return/{return_id}/confirm ───────────────────────────

func (h *Handler) ConfirmReturn(w http.ResponseWriter, r *http.Request) {
	returnID, err := parseReturnID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	result, err := h.svc.ConfirmReturn(username, returnID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "return record not found", "invoice not found":
			statusCode = http.StatusNotFound
		case "return already confirmed or in invalid status":
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GetReturnsByInvoice  GET /invoice/{invoice_id}/returns ──────────────────

func (h *Handler) GetReturnsByInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	result, err := h.svc.GetReturnsByInvoice(username, invoiceID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "invoice not found" {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}
	jsonResponse(w, map[string]interface{}{"returns": result}, http.StatusOK)
}

// ─── GetReturn  GET /return/{return_id} ──────────────────────────────────────

func (h *Handler) GetReturn(w http.ResponseWriter, r *http.Request) {
	returnID, err := parseReturnID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	result, err := h.svc.GetReturn(username, returnID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "return record not found" {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── DeleteReturn  DELETE /return/{return_id} ─────────────────────────────────

func (h *Handler) DeleteReturn(w http.ResponseWriter, r *http.Request) {
	returnID, err := parseReturnID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	if err := h.svc.DeleteReturn(username, returnID); err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "return record not found":
			statusCode = http.StatusNotFound
		case "cannot delete a return that is already confirmed or denied":
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message": "Return invoice deleted successfully",
	}, http.StatusOK)
}

// ─── parseReturnID helper ─────────────────────────────────────────────────────

func parseReturnID(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	return strconv.ParseInt(vars["return_id"], 10, 64)
}

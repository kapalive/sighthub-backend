package invoice

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	invSvc "sighthub-backend/internal/services/patient_service/invoice"
	pkgAuth "sighthub-backend/pkg/auth"
)

// ─── AddPatientPayment  POST /invoice/{invoice_id}/patient-payment ──────────

func (h *Handler) AddPatientPayment(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.PatientPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.AddPatientPayment(username, invoiceID, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found":
			statusCode = http.StatusNotFound
		case "payments only at invoice location", "invoice is finalized":
			statusCode = http.StatusForbidden
		case "invalid amount", "amount must be > 0", "adjustment amount cannot be zero":
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── PayWithCredit  POST /invoice/{invoice_id}/credit/payment ───────────────

func (h *Handler) PayWithCredit(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.CreditPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.PayWithCredit(username, invoiceID, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found":
			statusCode = http.StatusNotFound
		case "credit payments only at invoice location", "invoice is finalized":
			statusCode = http.StatusForbidden
		case "invoice has no due balance", "amount must be > 0", "cannot exceed invoice due", "not enough credit":
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── AddDiscount  POST /invoice/{invoice_id}/discount ───────────────────────

func (h *Handler) AddDiscount(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.DiscountInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	due, err := h.svc.AddDiscount(username, invoiceID, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		if msg == "invoice not found" {
			statusCode = http.StatusNotFound
		} else if msg == "invoice is finalized" {
			statusCode = http.StatusForbidden
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Discount updated", "due": due}, http.StatusOK)
}

// ─── GetInsurancePaymentTypes  GET /insurance-payment-types ─────────────────

func (h *Handler) GetInsurancePaymentTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.svc.GetInsurancePaymentTypes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, types, http.StatusOK)
}

// ─── AddInsurancePayment  POST /invoice/{invoice_id}/insurance-payment ──────

func (h *Handler) AddInsurancePayment(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.InsurancePaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.AddInsurancePayment(username, invoiceID, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found":
			statusCode = http.StatusNotFound
		case "invoice is finalized":
			statusCode = http.StatusForbidden
		case "no insurance policy attached", "invalid or inactive payment type", "amount must be > 0":
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GetInsurancePayments  GET /invoice/{invoice_id}/insurance-payments ─────

func (h *Handler) GetInsurancePayments(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetInsurancePayments(invoiceID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "invoice not found" {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── DeleteInsurancePayment  DELETE /invoice/{invoice_id}/insurance-payment/{payment_id} ─

func (h *Handler) DeleteInsurancePayment(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	paymentID, err := parsePaymentID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	due, err := h.svc.DeleteInsurancePayment(username, invoiceID, paymentID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found", "insurance payment not found":
			statusCode = http.StatusNotFound
		case "invoice is finalized":
			statusCode = http.StatusForbidden
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Insurance payment deleted", "due": due}, http.StatusOK)
}

// ─── GetPaymentHistory  GET /invoice/{invoice_id}/payment_history ───────────

func (h *Handler) GetPaymentHistory(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetPaymentHistory(invoiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GetCreditPayments  GET /{patient_id}/credit/payments ───────────────────

func (h *Handler) GetCreditPayments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	patientID, err := strconv.ParseInt(vars["patient_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid patient_id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	result, err := h.svc.GetCreditPayments(username, patientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GetCreditBalance  GET /credit_balance/{patient_id} ─────────────────────

func (h *Handler) GetCreditBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	patientID, err := strconv.ParseInt(vars["patient_id"], 10, 64)
	if err != nil {
		http.Error(w, "invalid patient_id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	result, err := h.svc.GetCreditBalance(username, patientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── DeletePayment  DELETE /invoice/{invoice_id}/payment/{payment_id} ───────

func (h *Handler) DeletePayment(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	paymentID, err := parsePaymentID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	due, err := h.svc.DeletePayment(username, invoiceID, paymentID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found", "payment not found":
			statusCode = http.StatusNotFound
		case "invoice is finalized", "wrong location":
			statusCode = http.StatusForbidden
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Payment deleted", "due": due}, http.StatusOK)
}

// ─── UpdatePayment  PUT /invoice/{invoice_id}/payment/{payment_id} ──────────

func (h *Handler) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	paymentID, err := parsePaymentID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.UpdatePaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.UpdatePayment(username, invoiceID, paymentID, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found", "payment not found":
			statusCode = http.StatusNotFound
		case "invoice is finalized", "credit updates only at invoice location":
			statusCode = http.StatusForbidden
		case "invalid amount", "not enough credit":
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── TransferCredit  POST /invoice/{invoice_id}/transfer-credit ─────────────

func (h *Handler) TransferCredit(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.TransferCreditInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.TransferCredit(username, invoiceID, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found", "payer patient not found":
			statusCode = http.StatusNotFound
		case "invoice is not in your current location", "invoice is finalized":
			statusCode = http.StatusForbidden
		case "invoice has no due balance", "amount must be > 0", "amount cannot exceed invoice due", "payer has insufficient credit balance":
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── parsePaymentID helper ───────────────────────────────────────────────────

func parsePaymentID(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	return strconv.ParseInt(vars["payment_id"], 10, 64)
}

package invoice

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	invSvc "sighthub-backend/internal/services/patient_service/invoice"
	pkgAuth "sighthub-backend/pkg/auth"
)

// ─── UpdateItem  PUT /invoice/{invoice_id}/item/{item_sale_id} ───────────────

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	itemSaleID, err := parseItemSaleID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.UpdateItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.UpdateItem(username, invoiceID, itemSaleID, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found", "invoice item not found":
			statusCode = http.StatusNotFound
		case "invoice is finalized":
			statusCode = http.StatusForbidden
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── DeleteItem  DELETE /invoice/{invoice_id}/item/{item_sale_id} ────────────

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	itemSaleID, err := parseItemSaleID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	if err := h.svc.DeleteItem(username, invoiceID, itemSaleID); err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found":
			statusCode = http.StatusNotFound
		case "invoice is finalized":
			statusCode = http.StatusForbidden
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, map[string]string{"message": "Item deleted"}, http.StatusOK)
}

// ─── SetLineBalance  PUT /invoice/{invoice_id}/item/{item_sale_id}/balance ───

func (h *Handler) SetLineBalance(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	itemSaleID, err := parseItemSaleID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.SetLineBalanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.SetLineBalance(username, invoiceID, itemSaleID, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found", "invoice item not found":
			statusCode = http.StatusNotFound
		case "invoice is finalized":
			statusCode = http.StatusForbidden
		case "provide 'pt_balance' and/or 'ins_balance'", "balances cannot be negative",
			"ins_balance cannot be negative", "pt_balance cannot be negative":
			statusCode = http.StatusBadRequest
		}
		// Range errors also 400
		if statusCode == http.StatusInternalServerError {
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── AddInsurancePolicy  POST /invoice/{invoice_id}/insurance/add ────────────

func (h *Handler) AddInsurancePolicy(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.AddInsurancePolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.InsuranceID == 0 {
		http.Error(w, "insurance_id is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.AddInsurancePolicy(username, invoiceID, input); err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found":
			statusCode = http.StatusNotFound
		case "invoice is finalized (locked) and cannot be updated":
			statusCode = http.StatusForbidden
		}
		if statusCode == http.StatusInternalServerError {
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, map[string]string{"message": "Insurance policy attached"}, http.StatusOK)
}

// ─── DeleteInsuranceFromInvoice  DELETE /invoice/insurance/{invoice_id} ──────

func (h *Handler) DeleteInsuranceFromInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	if err := h.svc.DeleteInsuranceFromInvoice(username, invoiceID); err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found":
			statusCode = http.StatusNotFound
		case "invoice is finalized (locked) and cannot be updated":
			statusCode = http.StatusForbidden
		case "no insurance policy associated with this invoice", "cannot delete insurance because insurance balance is not zero":
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, map[string]string{"message": "Insurance policy removed from invoice successfully"}, http.StatusOK)
}

// ─── AddGiftCard  POST /invoice/{invoice_id}/giftcard ───────────────────────

func (h *Handler) AddGiftCard(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.AddGiftCardInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.GiftCardID == 0 {
		http.Error(w, "gift card ID is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.AddGiftCard(username, invoiceID, input); err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found", "gift card not found":
			statusCode = http.StatusNotFound
		case "invoice is finalized (locked) and cannot be updated":
			statusCode = http.StatusForbidden
		case "not enough balance on gift card", "invalid gift card balance", "invalid data provided":
			statusCode = http.StatusBadRequest
		}
		if statusCode == http.StatusInternalServerError {
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, map[string]string{"message": "Gift card amount added successfully"}, http.StatusOK)
}

// ─── DeleteGiftCard  DELETE /invoice/gift-card/{invoice_id} ─────────────────

func (h *Handler) DeleteGiftCard(w http.ResponseWriter, r *http.Request) {
	invoiceID, err := parseInvoiceID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())

	if err := h.svc.DeleteGiftCard(username, invoiceID); err != nil {
		statusCode := http.StatusInternalServerError
		msg := err.Error()
		switch msg {
		case "invoice not found":
			statusCode = http.StatusNotFound
		case "no gift card associated with this invoice":
			statusCode = http.StatusBadRequest
		}
		http.Error(w, msg, statusCode)
		return
	}
	jsonResponse(w, map[string]string{"message": "Gift card removed from invoice successfully"}, http.StatusOK)
}

// ─── parseItemSaleID helper ───────────────────────────────────────────────────

func parseItemSaleID(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	return strconv.ParseInt(vars["item_sale_id"], 10, 64)
}

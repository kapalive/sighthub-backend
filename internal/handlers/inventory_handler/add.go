package inventory_handler

import (
	"net/http"

	invSvc "sighthub-backend/internal/services/inventory_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

// POST /add
func (h *Handler) AddInventoryItem(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	_ = username

	var input invSvc.AddInventoryInput
	if err := decodeJSON(r, &input); err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.AddInventoryItem(pkgAuth.UsernameFromContext(r.Context()), input)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 201, result)
}

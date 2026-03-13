package inventory_handler

import (
	"net/http"
)

// GET /preview-label — TODO: depends on label PDF generation (utils_printer)
func (h *Handler) PreviewLabel(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 501, map[string]string{"error": "label generation not yet implemented"})
}

// GET /print-label — TODO: depends on label PDF generation (utils_printer)
func (h *Handler) PrintLabel(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 501, map[string]string{"error": "label generation not yet implemented"})
}

// GET /print-labels-by-vendor — TODO: depends on label PDF generation (utils_printer)
func (h *Handler) PrintLabelsByVendor(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 501, map[string]string{"error": "label generation not yet implemented"})
}

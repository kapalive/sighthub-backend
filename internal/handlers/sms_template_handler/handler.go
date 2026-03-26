package sms_template_handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"sighthub-backend/internal/models/notifications"
)

type Handler struct{ db *gorm.DB }

func New(db *gorm.DB) *Handler { return &Handler{db: db} }

// GET /api/sms-templates
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var templates []notifications.SMSTemplate
	tx := h.db.Order("category, name")
	if category != "" {
		tx = tx.Where("category = ?", category)
	}
	tx.Find(&templates)

	items := make([]map[string]interface{}, 0, len(templates))
	for _, t := range templates {
		items = append(items, map[string]interface{}{
			"id":        t.IDSMSTemplate,
			"category":  t.Category,
			"name":      t.Name,
			"body":      t.Body,
			"is_system": t.IsSystem,
			"active":    t.Active,
		})
	}
	jsonResp(w, items, http.StatusOK)
}

// POST /api/sms-templates
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Category string `json:"category"`
		Name     string `json:"name"`
		Body     string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Category == "" || body.Name == "" || body.Body == "" {
		jsonErr(w, "category, name, body are required", http.StatusBadRequest)
		return
	}

	tpl := notifications.SMSTemplate{
		Category:  body.Category,
		Name:      body.Name,
		Body:      body.Body,
		IsSystem:  false,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.db.Create(&tpl).Error; err != nil {
		jsonErr(w, "failed to create: "+err.Error(), http.StatusConflict)
		return
	}
	jsonResp(w, map[string]interface{}{
		"id":       tpl.IDSMSTemplate,
		"category": tpl.Category,
		"name":     tpl.Name,
		"body":     tpl.Body,
	}, http.StatusCreated)
}

// PUT /api/sms-templates/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if id == 0 {
		jsonErr(w, "invalid id", http.StatusBadRequest)
		return
	}

	var tpl notifications.SMSTemplate
	if err := h.db.First(&tpl, id).Error; err != nil {
		jsonErr(w, "template not found", http.StatusNotFound)
		return
	}

	var body struct {
		Name   *string `json:"name"`
		Body   *string `json:"body"`
		Active *bool   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if body.Name != nil {
		if tpl.IsSystem {
			jsonErr(w, "cannot rename system template", http.StatusForbidden)
			return
		}
		tpl.Name = *body.Name
	}
	if body.Body != nil {
		tpl.Body = *body.Body
	}
	if body.Active != nil {
		tpl.Active = *body.Active
	}
	tpl.UpdatedAt = time.Now()

	h.db.Save(&tpl)
	jsonResp(w, map[string]interface{}{
		"id":       tpl.IDSMSTemplate,
		"category": tpl.Category,
		"name":     tpl.Name,
		"body":     tpl.Body,
		"active":   tpl.Active,
	}, http.StatusOK)
}

// DELETE /api/sms-templates/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if id == 0 {
		jsonErr(w, "invalid id", http.StatusBadRequest)
		return
	}

	var tpl notifications.SMSTemplate
	if err := h.db.First(&tpl, id).Error; err != nil {
		jsonErr(w, "template not found", http.StatusNotFound)
		return
	}
	if tpl.IsSystem {
		jsonErr(w, "cannot delete system template", http.StatusForbidden)
		return
	}

	h.db.Delete(&tpl)
	jsonResp(w, map[string]interface{}{"message": "deleted"}, http.StatusOK)
}

// GET /api/sms-templates/variables
func (h *Handler) GetVariables(w http.ResponseWriter, r *http.Request) {
	vars := []map[string]string{
		{"var": "{{.patient_name}}", "description": "Patient full name"},
		{"var": "{{.location}}", "description": "Store/location name"},
		{"var": "{{.location_phone}}", "description": "Store phone number"},
		{"var": "{{.location_address}}", "description": "Store address"},
		{"var": "{{.doctor}}", "description": "Doctor name (with Dr. prefix)"},
		{"var": "{{.date}}", "description": "Formatted date"},
		{"var": "{{.start_time}}", "description": "Appointment start time"},
		{"var": "{{.end_time}}", "description": "Appointment end time"},
		{"var": "{{.intake_url}}", "description": "Intake form URL (auto-generated)"},
		{"var": "{{.invoice_number}}", "description": "Invoice number"},
		{"var": "{{.ticket_number}}", "description": "Lab ticket number"},
	}
	jsonResp(w, vars, http.StatusOK)
}

func jsonResp(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

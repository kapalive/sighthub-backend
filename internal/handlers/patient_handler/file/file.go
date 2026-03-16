package file

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/patients"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ DB *gorm.DB }

func New(db *gorm.DB) *Handler { return &Handler{DB: db} }

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}

func cleanPath(path string) string {
	if idx := strings.Index(path, "/mnt/tank/data/"); idx != -1 {
		return strings.TrimSpace(path[idx+len("/mnt/tank/data/"):])
	}
	return path
}

func (h *Handler) getEmployee(username string) (*empModel.Employee, error) {
	var login authModel.EmployeeLogin
	if err := h.DB.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, err
	}
	var emp empModel.Employee
	if err := h.DB.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

// GET /api/patient/file?id_patient=X
func (h *Handler) GetPatientFiles(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id_patient")
	if idStr == "" {
		jsonError(w, "id_patient is required", http.StatusBadRequest)
		return
	}
	idPatient, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonError(w, "Invalid id_patient format", http.StatusBadRequest)
		return
	}

	var count int64
	h.DB.Model(&patients.Patient{}).Where("id_patient = ?", idPatient).Count(&count)
	if count == 0 {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	var docs []patients.DocumentsPatient
	if err := h.DB.Where("patient_id = ? AND is_hidden = false", idPatient).
		Order("created_time DESC").Find(&docs).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, 0, len(docs))
	for _, doc := range docs {
		result = append(result, map[string]interface{}{
			"id_file":       doc.IDDocumentsPatient,
			"file_name":     doc.FileName,
			"description":   doc.Description,
			"download_path": doc.FilePath,
		})
	}

	jsonOK(w, result)
}

// POST /api/patient/file
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDPatient int64  `json:"id_patient"`
		FilePath  string `json:"file_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.IDPatient == 0 {
		jsonError(w, "id_patient is required", http.StatusBadRequest)
		return
	}
	if input.FilePath == "" {
		jsonError(w, "file_path is required", http.StatusBadRequest)
		return
	}

	var count int64
	h.DB.Model(&patients.Patient{}).Where("id_patient = ?", input.IDPatient).Count(&count)
	if count == 0 {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	emp, err := h.getEmployee(username)
	if err != nil {
		jsonError(w, "Employee not found", http.StatusNotFound)
		return
	}

	cleanedPath := cleanPath(input.FilePath)
	empID := int64(emp.IDEmployee)
	doc := patients.DocumentsPatient{
		PatientID: input.IDPatient,
		FileName:  fmt.Sprintf("doc-%d.pdf", input.IDPatient),
		FilePath:  cleanedPath,
		CreatedBy: &empID,
	}
	if err := h.DB.Create(&doc).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]interface{}{
		"status":  "success",
		"message": "File path processed and saved successfully.",
		"id_file": doc.IDDocumentsPatient,
	})
}

// GET /api/patient/file/{id_file}
func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {
	idFile, err := strconv.ParseInt(mux.Vars(r)["id_file"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid id_file", http.StatusBadRequest)
		return
	}

	var doc patients.DocumentsPatient
	if err := h.DB.First(&doc, idFile).Error; err != nil {
		jsonError(w, "File not found", http.StatusNotFound)
		return
	}

	createdBy := "Unknown"
	if doc.CreatedBy != nil {
		var emp empModel.Employee
		if err := h.DB.First(&emp, *doc.CreatedBy).Error; err == nil {
			createdBy = fmt.Sprintf("%s %s", emp.FirstName, emp.LastName)
		}
	}

	var createdTime interface{}
	if doc.CreatedTime != nil {
		createdTime = doc.CreatedTime.Format("2006-01-02 15:04:05")
	}

	jsonOK(w, map[string]interface{}{
		"name":         doc.FileName,
		"description":  doc.Description,
		"id_file":      doc.IDDocumentsPatient,
		"created_by":   createdBy,
		"created_time": createdTime,
		"path":         doc.FilePath,
	})
}

// PUT /api/patient/file/{id_file}
func (h *Handler) UpdateFile(w http.ResponseWriter, r *http.Request) {
	idFile, err := strconv.ParseInt(mux.Vars(r)["id_file"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid id_file", http.StatusBadRequest)
		return
	}

	var doc patients.DocumentsPatient
	if err := h.DB.First(&doc, idFile).Error; err != nil {
		jsonError(w, "File not found", http.StatusNotFound)
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if v, ok := input["file_name"].(string); ok {
		doc.FileName = v
	}
	if v, ok := input["description"].(string); ok {
		doc.Description = &v
	}
	if v, ok := input["path"].(string); ok {
		doc.FilePath = v
	}

	if err := h.DB.Save(&doc).Error; err != nil {
		jsonError(w, "An error occurred while updating file metadata", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"status":  "success",
		"message": "File metadata updated successfully.",
	})
}

// DELETE /api/patient/file/{id_file}
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	idFile, err := strconv.ParseInt(mux.Vars(r)["id_file"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid id_file", http.StatusBadRequest)
		return
	}

	var doc patients.DocumentsPatient
	if err := h.DB.First(&doc, idFile).Error; err != nil {
		jsonError(w, "Document not found", http.StatusNotFound)
		return
	}

	// Soft delete: mark as hidden
	doc.IsHidden = true
	if err := h.DB.Save(&doc).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"message": "File successfully deleted"})
}

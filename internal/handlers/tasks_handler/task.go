package tasks_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	pkgAuth "sighthub-backend/pkg/auth"
	svc "sighthub-backend/internal/services/task_service"
)

type Handler struct {
	svc *svc.Service
	db  *gorm.DB
}

func New(s *svc.Service, db *gorm.DB) *Handler { return &Handler{svc: s, db: db} }

// ─── helpers ──────────────────────────────────────────────────────────────────

func jsonOK(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// getEmployee resolves the current JWT user to an Employee record.
func (h *Handler) getEmployee(r *http.Request) (*svc.Service, int, int64, error) {
	username := pkgAuth.UsernameFromContext(r.Context())
	emp, _, err := h.svc.GetEmployeeByUsername(username)
	if err != nil || emp == nil {
		return h.svc, 0, 0, err
	}
	var locID int64
	if emp.LocationID != nil {
		locID = *emp.LocationID
	}
	return h.svc, emp.IDEmployee, locID, nil
}

// ─── GET /tasks ───────────────────────────────────────────────────────────────

func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	_, empID, _, err := h.getEmployee(r)
	if err != nil || empID == 0 {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	q := r.URL.Query()
	mode := q.Get("mode")
	if mode == "" {
		mode = "for_me"
	}
	statuses := q["status"]

	var dateStart, dateEnd *string
	if v := q.Get("date_start"); v != "" {
		dateStart = &v
	}
	if v := q.Get("date_end"); v != "" {
		dateEnd = &v
	}

	tasks, err := h.svc.GetTasks(svc.GetTasksFilters{
		Mode:       mode,
		Statuses:   statuses,
		DateStart:  dateStart,
		DateEnd:    dateEnd,
		EmployeeID: empID,
	})
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, map[string]interface{}{"tasks": tasks})
}

// ─── POST /tasks ──────────────────────────────────────────────────────────────

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	_, empID, _, err := h.getEmployee(r)
	if err != nil || empID == 0 {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	var input svc.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}
	input.CreatorID = empID

	task, err := h.svc.CreateTask(input)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]interface{}{"message": "Task created", "task": task})
}

// ─── GET /tasks/{id} ──────────────────────────────────────────────────────────

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id_task"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	detail, err := h.svc.GetTask(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if detail == nil {
		jsonError(w, "Task not found", http.StatusNotFound)
		return
	}
	jsonOK(w, detail)
}

// ─── PUT /tasks/{id} ─────────────────────────────────────────────────────────

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id_task"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var input svc.UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "No input provided", http.StatusBadRequest)
		return
	}

	task, err := h.svc.UpdateTask(id, input)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if task == nil {
		jsonError(w, "Task not found", http.StatusNotFound)
		return
	}
	jsonOK(w, map[string]interface{}{"message": "Task updated", "task": task})
}

// ─── DELETE /tasks/{id} ───────────────────────────────────────────────────────

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id_task"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteTask(id); err != nil {
		if err.Error() == "not found" {
			jsonError(w, "Task not found", http.StatusNotFound)
		} else {
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	jsonOK(w, map[string]string{"message": "Task " + strconv.FormatInt(id, 10) + " deleted"})
}

// ─── DELETE /tasks/resource/{id} ─────────────────────────────────────────────

func (h *Handler) DeleteTaskResource(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id_tasks_resource"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteTaskResource(id); err != nil {
		if err.Error() == "not found" {
			jsonError(w, "Resource not found", http.StatusNotFound)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonOK(w, map[string]string{"message": "Resource " + strconv.FormatInt(id, 10) + " deleted successfully"})
}

// ─── GET /tasks/invoice/search ────────────────────────────────────────────────

func (h *Handler) SearchInvoice(w http.ResponseWriter, r *http.Request) {
	_, empID, locID, err := h.getEmployee(r)
	if err != nil || empID == 0 {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}
	if locID == 0 {
		jsonError(w, "Employee has no location", http.StatusNotFound)
		return
	}

	query := r.URL.Query().Get("number_invoice")
	results, err := h.svc.SearchInvoice(query, locID)
	if err != nil {
		msg := err.Error()
		if msg == "invoice number not provided" {
			jsonError(w, msg, http.StatusBadRequest)
		} else if msg == "no invoices found" {
			jsonError(w, msg, http.StatusNotFound)
		} else if msg == "no accessible invoices found" {
			jsonError(w, msg, http.StatusForbidden)
		} else {
			jsonError(w, msg, http.StatusInternalServerError)
		}
		return
	}
	jsonOK(w, map[string]interface{}{"invoices": results})
}

// ─── GET /tasks/employees ─────────────────────────────────────────────────────

func (h *Handler) GetEmployees(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetEmployees()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, list)
}

// ─── POST /tasks/{id}/comments ────────────────────────────────────────────────

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(mux.Vars(r)["id_task"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	_, empID, _, err := h.getEmployee(r)
	if err != nil || empID == 0 {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	var body struct {
		Message string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if body.Message == "" {
		jsonError(w, "message is required", http.StatusBadRequest)
		return
	}

	result, err := h.svc.CreateComment(taskID, empID, body.Message)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result == nil {
		jsonError(w, "Task not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, result)
}

// ─── PUT /tasks/comments/{id} ─────────────────────────────────────────────────

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := strconv.ParseInt(mux.Vars(r)["id_task_comment"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	_, empID, _, err := h.getEmployee(r)
	if err != nil || empID == 0 {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	var body struct {
		Message string `json:"message"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if body.Message == "" {
		jsonError(w, "message is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateComment(commentID, empID, body.Message); err != nil {
		if err.Error() == "not found" {
			jsonError(w, "Comment not found", http.StatusNotFound)
		} else if err.Error() == "forbidden" {
			jsonError(w, "You can only edit your own comments", http.StatusForbidden)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonOK(w, map[string]string{"message": "Comment updated"})
}

// ─── DELETE /tasks/comments/{id} ──────────────────────────────────────────────

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := strconv.ParseInt(mux.Vars(r)["id_task_comment"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	_, empID, _, err := h.getEmployee(r)
	if err != nil || empID == 0 {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	if err := h.svc.DeleteComment(commentID, empID); err != nil {
		if err.Error() == "not found" {
			jsonError(w, "Comment not found", http.StatusNotFound)
		} else if err.Error() == "forbidden" {
			jsonError(w, "You can only delete your own comments", http.StatusForbidden)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonOK(w, map[string]string{"message": "Comment deleted"})
}

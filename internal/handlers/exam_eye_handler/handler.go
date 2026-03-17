package exam_eye_handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	eeSvc "sighthub-backend/internal/services/exam_eye_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

// flexInt64 unmarshals from both JSON number (42) and JSON string ("42").
type flexInt64 int64

func (f *flexInt64) UnmarshalJSON(b []byte) error {
	var n int64
	if err := json.Unmarshal(b, &n); err == nil {
		*f = flexInt64(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %q as int64", s)
		}
		*f = flexInt64(n)
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into int64", string(b))
}

type Handler struct{ svc *eeSvc.Service }

func New(svc *eeSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func parsePathInt64(r *http.Request, key string) (int64, bool) {
	v, ok := mux.Vars(r)[key]
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	return n, err == nil
}

// GET /exam-types
func (h *Handler) GetExamTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.svc.GetExamTypes()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{"exam_types": types})
}

// POST /new
func (h *Handler) StartNewExam(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var body struct {
		PatientID     flexInt64 `json:"patient_id"`
		EyeExamTypeID flexInt64 `json:"eye_exam_type_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PatientID == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "patient_id is required"})
		return
	}

	result, err := h.svc.StartNewExam(username, int64(body.PatientID), int64(body.EyeExamTypeID))
	if err != nil {
		switch err.Error() {
		case "employee or location not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "patient not found", "exam type not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message":     "Exam started successfully",
		"exam_id":     result.ExamID,
		"exam_name":   result.ExamName,
		"patient_age": result.PatientAge,
		"sub_routes":  result.SubRoutes,
	})
}

// PUT /submit/{exam_id}
func (h *Handler) SubmitExam(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	if err := h.svc.SubmitExam(username, examID); err != nil {
		switch err.Error() {
		case "employee or location not found", "exam not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "cannot update a completed exam", "you are not authorized to update this exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Exam completed successfully"})
}

// GET /{id_exam}
func (h *Handler) GetExamDetails(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "id_exam")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid id_exam"})
		return
	}

	result, err := h.svc.GetExamDetails(examID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"message": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// DELETE /cancel/{id_exam}
func (h *Handler) CancelExam(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "id_exam")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid id_exam"})
		return
	}

	if err := h.svc.CancelExam(username, examID); err != nil {
		switch err.Error() {
		case "employee or location not found", "exam not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "you are not authorized to cancel this exam", "cannot cancel a completed exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Exam canceled successfully"})
}

// PUT /unlock/{exam_id}
func (h *Handler) UnlockExam(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	if err := h.svc.UnlockExam(username, examID); err != nil {
		switch err.Error() {
		case "employee or location not found", "exam not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "exam is already unlocked", "you are not authorized to unlock this exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Exam unlocked successfully"})
}

// GET /notes
func (h *Handler) GetNotes(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	objectName := r.URL.Query().Get("object")
	fieldName := r.URL.Query().Get("field")
	if objectName == "" || fieldName == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "both 'object' and 'field' parameters are required"})
		return
	}

	notes, err := h.svc.GetNotes(username, objectName, fieldName)
	if err != nil {
		switch err.Error() {
		case "employee or location not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			if len(err.Error()) > 8 && err.Error()[:8] == "table '" || len(err.Error()) > 9 && err.Error()[:9] == "column '" {
				jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			} else {
				jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		return
	}
	jsonResponse(w, http.StatusOK, notes)
}

// POST /notes
func (h *Handler) AddNote(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var body struct {
		Object   string `json:"object"`
		Field    string `json:"field"`
		Note     string `json:"note"`
		Priority int    `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if body.Object == "" || body.Field == "" || body.Note == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "parameters 'object', 'field', and 'note' are required"})
		return
	}

	note, err := h.svc.AddNote(username, body.Object, body.Field, body.Note, body.Priority)
	if err != nil {
		switch err.Error() {
		case "employee or location not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]interface{}{"message": "Note added successfully", "note": note})
}

// PUT /notes
func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NoteID   *int64 `json:"note_id"`
		Priority *int   `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.NoteID == nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "'note_id' is required"})
		return
	}

	result, err := h.svc.UpdateNote(*body.NoteID, body.Priority)
	if err != nil {
		if err.Error() == "note not found" {
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		} else {
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	// result may contain swap fields or just the note
	if _, hasMsg := result["message"]; hasMsg {
		jsonResponse(w, http.StatusOK, result)
	} else {
		jsonResponse(w, http.StatusOK, map[string]interface{}{"message": "Note updated successfully", "updated_note": result})
	}
}

// DELETE /notes
func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	noteIDStr := r.URL.Query().Get("note_id")
	if noteIDStr == "" {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "'note_id' is required as a query parameter"})
		return
	}
	noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid note_id"})
		return
	}

	if err := h.svc.DeleteNote(username, noteID); err != nil {
		switch err.Error() {
		case "employee or location not found", "note not found", "associated note document not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "you do not have permission to delete this note":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Note deleted successfully"})
}

// PUT /change-type/{exam_id}
func (h *Handler) ChangeExamType(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var body struct {
		EyeExamTypeID int64 `json:"eye_exam_type_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.EyeExamTypeID == 0 {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "field 'eye_exam_type_id' is required"})
		return
	}

	result, err := h.svc.ChangeExamType(username, examID, body.EyeExamTypeID)
	if err != nil {
		switch err.Error() {
		case "employee or location not found", "exam not found", "exam type not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "you are not authorized to change this exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case "cannot change type of a completed exam":
			jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

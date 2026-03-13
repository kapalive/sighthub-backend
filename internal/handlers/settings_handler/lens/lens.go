package lens

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	svc "sighthub-backend/internal/services/settings_service"
)

type Handler struct{ svc *svc.Service }

func New(s *svc.Service) *Handler { return &Handler{svc: s} }

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func pathID(r *http.Request) int {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	return id
}

// ── Lens Types ─────────────────────────────────────────────────────────────

func (h *Handler) ListTypes(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListLensTypes()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateType(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TypeName    string `json:"type_name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateLensType(body.TypeName, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateType(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.UpdateLensType(pathID(r), body)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) DeleteType(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLensType(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Lens Materials ─────────────────────────────────────────────────────────

func (h *Handler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListLensMaterials()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MaterialName string   `json:"material_name"`
		Index        *float64 `json:"index"`
		Description  *string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateLensMaterial(body.MaterialName, body.Index, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateLensMaterial(pathID(r), body); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLensMaterial(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Lens Special Features ──────────────────────────────────────────────────

func (h *Handler) ListSpecialFeatures(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListLensSpecialFeatures()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateSpecialFeature(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FeatureName string `json:"feature_name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateLensSpecialFeature(body.FeatureName, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateSpecialFeature(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateLensSpecialFeature(pathID(r), body); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteSpecialFeature(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLensSpecialFeature(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Lens Series ────────────────────────────────────────────────────────────

func (h *Handler) ListSeries(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListLensSeries()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateSeries(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SeriesName  string `json:"series_name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateLensSeries(body.SeriesName, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateSeries(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateLensSeries(pathID(r), body); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteSeries(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLensSeries(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── VCodes ─────────────────────────────────────────────────────────────────

func (h *Handler) ListVCodes(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListVCodes()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateVCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateVCode(body.Code, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateVCode(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateVCode(pathID(r), body); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteVCode(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteVCode(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Lens Styles ────────────────────────────────────────────────────────────

func (h *Handler) ListStyles(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListLensStyles()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateStyle(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StyleName   string `json:"style_name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateLensStyle(body.StyleName, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateStyle(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateLensStyle(pathID(r), body); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteStyle(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLensStyle(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Tint Colors ────────────────────────────────────────────────────────────

func (h *Handler) ListTintColors(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListLensTintColors()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateTintColor(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"lens_tint_color_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateLensTintColor(body.Name)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateTintColor(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"lens_tint_color_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateLensTintColor(pathID(r), body.Name); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteTintColor(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLensTintColor(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Sample Colors ──────────────────────────────────────────────────────────

func (h *Handler) ListSampleColors(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListLensSampleColors()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateSampleColor(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"lens_sample_color_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateLensSampleColor(body.Name)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateSampleColor(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"lens_sample_color_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateLensSampleColor(pathID(r), body.Name); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteSampleColor(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLensSampleColor(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Safety Thickness ───────────────────────────────────────────────────────

func (h *Handler) ListSafetyThickness(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListLensSafetyThickness()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateSafetyThickness(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"safety_thickness_name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateLensSafetyThickness(body.Name, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateSafetyThickness(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateLensSafetyThickness(pathID(r), body); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteSafetyThickness(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLensSafetyThickness(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Bevels ─────────────────────────────────────────────────────────────────

func (h *Handler) ListBevels(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListLensBevels()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateBevel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"lens_bevel_name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateLensBevel(body.Name, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateBevel(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateLensBevel(pathID(r), body); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteBevel(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLensBevel(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Lens Statuses ──────────────────────────────────────────────────────────

func (h *Handler) ListLensStatuses(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListLensStatuses()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateLensStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StatusName string `json:"status_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateLensStatus(body.StatusName)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateLensStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StatusName string `json:"status_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateLensStatus(pathID(r), body.StatusName); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteLensStatus(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLensStatus(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

package insurance

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	insModel "sighthub-backend/internal/models/insurance"
	"sighthub-backend/internal/models/patients"
)

type Handler struct{ DB *gorm.DB }

func New(db *gorm.DB) *Handler { return &Handler{DB: db} }

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func jsonCreated(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}

func cleanPath(v string) *string {
	if v == "" {
		return nil
	}
	if idx := strings.Index(v, "/mnt/tank/data/"); idx != -1 {
		v = strings.TrimSpace(v[idx+len("/mnt/tank/data/"):])
	}
	if v == "" {
		return nil
	}
	return &v
}

// GET /api/patient/insurance?id_patient=X
func (h *Handler) GetInsurancePolicies(w http.ResponseWriter, r *http.Request) {
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

	var holders []patients.InsuranceHolderPatients
	if err := h.DB.Where("patient_id = ?", idPatient).Find(&holders).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, 0, len(holders))
	for _, holder := range holders {
		var policy insModel.InsurancePolicy
		if err := h.DB.
			Preload("InsuranceCompany").
			Preload("InsuranceCoverageType").
			First(&policy, holder.InsurancePolicyID).Error; err != nil {
			continue
		}

		companyName := "Unknown Provider"
		var companyID interface{}
		if policy.InsuranceCompany != nil {
			companyName = policy.InsuranceCompany.CompanyName
			companyID = policy.InsuranceCompany.IDInsuranceCompany
		}

		var coverageName interface{}
		if policy.InsuranceCoverageType != nil {
			coverageName = policy.InsuranceCoverageType.CoverageName
		}

		position := "Unknown Position"
		if holder.Position != nil {
			position = *holder.Position
		}

		result = append(result, map[string]interface{}{
			"id_insurance":               policy.IDInsurancePolicy,
			"insurance_company":          companyName,
			"insurance_company_id":       companyID,
			"coverage_name":              coverageName,
			"id_insurance_coverage_type": policy.InsuranceCoverageTypeID,
			"group_number":               policy.GroupNumber,
			"member_number":              holder.MemberNumber,
			"coverage_details":           policy.CoverageDetails,
			"specify":                    policy.Specify,
			"active":                     policy.Active,
			"front_photo":                policy.FrontPhoto,
			"back_photo":                 policy.BackPhoto,
			"holder_type":                holder.HolderType,
			"position":                   position,
		})
	}

	jsonOK(w, result)
}

// POST /api/patient/insurance
func (h *Handler) CreateInsurancePolicy(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// id_patient
	idPatientRaw, ok := data["id_patient"]
	if !ok {
		jsonError(w, "id_patient is required", http.StatusBadRequest)
		return
	}
	idPatient, err := toInt64(idPatientRaw)
	if err != nil {
		jsonError(w, "Invalid id_patient format", http.StatusBadRequest)
		return
	}
	var patCount int64
	h.DB.Model(&patients.Patient{}).Where("id_patient = ?", idPatient).Count(&patCount)
	if patCount == 0 {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	// member_number + insurance_company_id required
	memberNumber, _ := data["member_number"].(string)
	if memberNumber == "" {
		jsonError(w, "member_number and insurance_company_id are required", http.StatusBadRequest)
		return
	}
	companyIDRaw, ok := data["insurance_company_id"]
	if !ok {
		jsonError(w, "member_number and insurance_company_id are required", http.StatusBadRequest)
		return
	}
	companyID, err := toInt64(companyIDRaw)
	if err != nil {
		jsonError(w, "Invalid insurance_company_id format", http.StatusBadRequest)
		return
	}

	var company insModel.InsuranceCompany
	if err := h.DB.First(&company, companyID).Error; err != nil {
		jsonError(w, "Insurance company not found", http.StatusNotFound)
		return
	}

	// Coverage type from request
	var coverageID *int
	if covRaw, ok := data["insurance_coverage_type_id"]; ok && covRaw != nil {
		covID, err := toInt64(covRaw)
		if err == nil && covID > 0 {
			id := int(covID)
			var cov insModel.InsuranceCoverageType
			if h.DB.First(&cov, id).Error == nil {
				coverageID = &id
			}
		}
	}

	active := true
	if v, ok := data["active"].(bool); ok {
		active = v
	}

	frontPhoto := cleanPath(toString(data["front_photo"]))
	backPhoto := cleanPath(toString(data["back_photo"]))

	var specify *string
	if v, ok := data["specify"].(string); ok {
		v = strings.TrimSpace(v)
		if v != "" {
			specify = &v
		}
	}

	groupNumber := toStrPtr(data["group_number"])
	coverageDetails := toStrPtr(data["coverage_details"])

	policy := insModel.InsurancePolicy{
		GroupNumber:             groupNumber,
		CoverageDetails:        coverageDetails,
		Specify:                specify,
		InsuranceCompanyID:     int(companyID),
		InsuranceCoverageTypeID: coverageID,
		Active:                 active,
		FrontPhoto:             frontPhoto,
		BackPhoto:              backPhoto,
	}

	tx := h.DB.Begin()
	if err := tx.Create(&policy).Error; err != nil {
		tx.Rollback()
		jsonError(w, "An error occurred while adding the insurance", http.StatusInternalServerError)
		return
	}

	primaryPos := "Primary"
	selfHolder := patients.InsuranceHolderPatients{
		InsurancePolicyID: policy.IDInsurancePolicy,
		PatientID:         idPatient,
		MemberNumber:      &memberNumber,
		HolderType:        "Self",
		Position:          &primaryPos,
	}
	if err := tx.Create(&selfHolder).Error; err != nil {
		tx.Rollback()
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If policy_holder != Self, add secondary holder
	policyHolder, _ := data["policy_holder"].(string)
	if policyHolder != "" && policyHolder != "Self" {
		holderIDRaw, ok := data["holder_id"]
		if !ok {
			tx.Rollback()
			jsonError(w, "holder_id is required if policy_holder != Self", http.StatusBadRequest)
			return
		}
		holderID, err := toInt64(holderIDRaw)
		if err != nil {
			tx.Rollback()
			jsonError(w, "Invalid holder_id format", http.StatusBadRequest)
			return
		}
		var hCount int64
		tx.Model(&patients.Patient{}).Where("id_patient = ?", holderID).Count(&hCount)
		if hCount == 0 {
			tx.Rollback()
			jsonError(w, "Holder patient not found", http.StatusNotFound)
			return
		}

		secPos := "Secondary"
		secHolder := patients.InsuranceHolderPatients{
			InsurancePolicyID: policy.IDInsurancePolicy,
			PatientID:         holderID,
			MemberNumber:      &memberNumber,
			HolderType:        policyHolder,
			Position:          &secPos,
		}
		if err := tx.Create(&secHolder).Error; err != nil {
			tx.Rollback()
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()

	jsonCreated(w, map[string]interface{}{
		"status":    "success",
		"message":   "Insurance added successfully.",
		"insurance": policy.ToMap(),
	})
}

// GET /api/patient/insurance/coverage_types
func (h *Handler) GetCoverageTypes(w http.ResponseWriter, r *http.Request) {
	var types []insModel.InsuranceCoverageType
	h.DB.Find(&types)

	result := make([]map[string]interface{}, 0, len(types))
	for _, t := range types {
		result = append(result, t.ToMap())
	}
	jsonOK(w, result)
}

// GET /api/patient/insurance/companies
func (h *Handler) GetCompanies(w http.ResponseWriter, r *http.Request) {
	var companies []insModel.InsuranceCompany
	h.DB.Order("company_name ASC").Find(&companies)

	result := make([]map[string]interface{}, 0, len(companies))
	for _, c := range companies {
		result = append(result, map[string]interface{}{
			"id_insurance_company": c.IDInsuranceCompany,
			"company_name":         c.CompanyName,
		})
	}
	jsonOK(w, result)
}

// GET /api/patient/insurance/{id_insurance}/patient/{id_patient}
func (h *Handler) GetInsurancePolicyByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idInsurance, err := strconv.ParseInt(vars["id_insurance"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid id_insurance", http.StatusBadRequest)
		return
	}
	idPatient, err := strconv.ParseInt(vars["id_patient"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid id_patient", http.StatusBadRequest)
		return
	}

	var policy insModel.InsurancePolicy
	if err := h.DB.
		Preload("InsuranceCompany").
		Preload("InsuranceCoverageType").
		First(&policy, idInsurance).Error; err != nil {
		jsonError(w, "Insurance policy not found", http.StatusNotFound)
		return
	}

	company := policy.InsuranceCompany
	companyName := "Unknown Provider"
	if company != nil {
		companyName = company.CompanyName
	}

	policyCov := policy.InsuranceCoverageType

	// Get all holders for this policy
	var holders []patients.InsuranceHolderPatients
	h.DB.Where("insurance_policy_id = ?", idInsurance).Find(&holders)

	holdersList := make([]map[string]interface{}, 0)
	var currentHolder map[string]interface{}

	for _, hld := range holders {
		var pat patients.Patient
		h.DB.First(&pat, hld.PatientID)

		hData := map[string]interface{}{
			"id_holder":     hld.IDInsuranceHolderPatients,
			"patient_id":    hld.PatientID,
			"first_name":    pat.FirstName,
			"last_name":     pat.LastName,
			"holder_type":   hld.HolderType,
			"position":      hld.Position,
			"member_number": hld.MemberNumber,
		}

		if hld.PatientID == idPatient {
			currentHolder = hData
		} else {
			holdersList = append(holdersList, hData)
		}
	}

	if currentHolder == nil {
		currentHolder = map[string]interface{}{}
	}

	var policyCovID, policyCovName interface{}
	if policyCov != nil {
		policyCovID = policyCov.IDInsuranceCoverageType
		policyCovName = policyCov.CoverageName
	}

	var companyIDVal interface{}
	if company != nil {
		companyIDVal = company.IDInsuranceCompany
	}

	jsonOK(w, map[string]interface{}{
		"insurance_info": map[string]interface{}{
			"id_insurance":                  policy.IDInsurancePolicy,
			"insurance_company_id":          companyIDVal,
			"company_name":                  companyName,
			"group_number":                  policy.GroupNumber,
			"coverage_details":              policy.CoverageDetails,
			"specify":                       policy.Specify,
			"active":                        policy.Active,
			"front_photo":                   policy.FrontPhoto,
			"back_photo":                    policy.BackPhoto,
			"insurance_coverage_type_id":    policyCovID,
			"insurance_coverage_type_name":  policyCovName,
		},
		"current_holder": currentHolder,
		"holders":        holdersList,
	})
}

// PUT /api/patient/insurance/{id_insurance}
func (h *Handler) UpdateInsurancePolicy(w http.ResponseWriter, r *http.Request) {
	idInsurance, err := strconv.ParseInt(mux.Vars(r)["id_insurance"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid id_insurance", http.StatusBadRequest)
		return
	}

	var policy insModel.InsurancePolicy
	if err := h.DB.First(&policy, idInsurance).Error; err != nil {
		jsonError(w, "Insurance policy not found", http.StatusNotFound)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if v, ok := data["group_number"]; ok {
		policy.GroupNumber = toStrPtr(v)
	}
	if v, ok := data["coverage_details"]; ok {
		policy.CoverageDetails = toStrPtr(v)
	}
	if v, ok := data["active"].(bool); ok {
		policy.Active = v
	}
	if v, ok := data["specify"]; ok {
		s := toString(v)
		s = strings.TrimSpace(s)
		if len(s) > 255 {
			jsonError(w, "specify must be <= 255 characters", http.StatusBadRequest)
			return
		}
		if s == "" {
			policy.Specify = nil
		} else {
			policy.Specify = &s
		}
	}

	// Check for coverage override
	coverageOverride := false
	var coverageKey string
	for _, k := range []string{"insurance_coverage_type_id", "id_insurance_coverage_type", "id_type_insurance_policy"} {
		if _, ok := data[k]; ok {
			coverageOverride = true
			coverageKey = k
			break
		}
	}

	if v, ok := data["insurance_company_id"]; ok {
		compID, err := toInt64(v)
		if err != nil {
			jsonError(w, "Invalid insurance_company_id format", http.StatusBadRequest)
			return
		}
		var comp insModel.InsuranceCompany
		if err := h.DB.First(&comp, compID).Error; err != nil {
			jsonError(w, "Insurance company not found", http.StatusNotFound)
			return
		}
		policy.InsuranceCompanyID = int(compID)
	}

	if coverageOverride {
		covVal := data[coverageKey]
		if covVal == nil || covVal == "" {
			policy.InsuranceCoverageTypeID = nil
		} else {
			covID, err := toInt64(covVal)
			if err != nil {
				jsonError(w, "Invalid coverage id format", http.StatusBadRequest)
				return
			}
			var cov insModel.InsuranceCoverageType
			if err := h.DB.First(&cov, covID).Error; err != nil {
				jsonError(w, "Coverage type not found", http.StatusNotFound)
				return
			}
			id := int(covID)
			policy.InsuranceCoverageTypeID = &id
		}
	}

	if v, ok := data["front_photo"]; ok {
		policy.FrontPhoto = cleanPath(toString(v))
	}
	if v, ok := data["back_photo"]; ok {
		policy.BackPhoto = cleanPath(toString(v))
	}

	if err := h.DB.Save(&policy).Error; err != nil {
		jsonError(w, "An error occurred while updating the insurance policy", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"status":    "success",
		"message":   "Insurance policy updated successfully.",
		"insurance": policy.ToMap(),
	})
}

// POST /api/patient/insurance/{id_insurance}/holders
func (h *Handler) AddHolder(w http.ResponseWriter, r *http.Request) {
	idInsurance, err := strconv.ParseInt(mux.Vars(r)["id_insurance"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid id_insurance", http.StatusBadRequest)
		return
	}

	var count int64
	h.DB.Model(&insModel.InsurancePolicy{}).Where("id_insurance_policy = ?", idInsurance).Count(&count)
	if count == 0 {
		jsonError(w, "Insurance policy not found", http.StatusNotFound)
		return
	}

	var input struct {
		IDPatient    int64  `json:"id_patient"`
		HolderType   string `json:"holder_type"`
		Position     string `json:"position"`
		MemberNumber string `json:"member_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.IDPatient == 0 {
		jsonError(w, "patient_id is required", http.StatusBadRequest)
		return
	}
	if input.HolderType == "" {
		jsonError(w, "holder_type is required", http.StatusBadRequest)
		return
	}
	if input.Position == "" {
		jsonError(w, "position is required", http.StatusBadRequest)
		return
	}
	if input.MemberNumber == "" {
		jsonError(w, "member_number is required", http.StatusBadRequest)
		return
	}

	h.DB.Model(&patients.Patient{}).Where("id_patient = ?", input.IDPatient).Count(&count)
	if count == 0 {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	// Check not already a holder
	h.DB.Model(&patients.InsuranceHolderPatients{}).
		Where("insurance_policy_id = ? AND patient_id = ?", idInsurance, input.IDPatient).Count(&count)
	if count > 0 {
		jsonError(w, "Patient is already a holder of this insurance policy.", http.StatusBadRequest)
		return
	}

	holder := patients.InsuranceHolderPatients{
		InsurancePolicyID: idInsurance,
		PatientID:         input.IDPatient,
		HolderType:        input.HolderType,
		Position:          &input.Position,
		MemberNumber:      &input.MemberNumber,
	}
	if err := h.DB.Create(&holder).Error; err != nil {
		jsonError(w, "An error occurred while adding the insurance holder", http.StatusInternalServerError)
		return
	}

	jsonCreated(w, map[string]interface{}{
		"status":  "success",
		"message": "Holder added successfully to the insurance policy.",
	})
}

// PUT /api/patient/insurance/{id_insurance}/holders/{patient_id}
func (h *Handler) UpdateHolder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idInsurance, _ := strconv.ParseInt(vars["id_insurance"], 10, 64)
	patientID, _ := strconv.ParseInt(vars["patient_id"], 10, 64)

	var count int64
	h.DB.Model(&insModel.InsurancePolicy{}).Where("id_insurance_policy = ?", idInsurance).Count(&count)
	if count == 0 {
		jsonError(w, "Insurance policy not found", http.StatusNotFound)
		return
	}

	var holder patients.InsuranceHolderPatients
	if err := h.DB.Where("insurance_policy_id = ? AND patient_id = ?", idInsurance, patientID).
		First(&holder).Error; err != nil {
		jsonError(w, "Holder not found for this insurance policy", http.StatusNotFound)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if v, ok := data["holder_type"].(string); ok {
		holder.HolderType = v
	}
	if v, ok := data["position"].(string); ok {
		holder.Position = &v
	}
	if v, ok := data["member_number"].(string); ok {
		holder.MemberNumber = &v
	}

	if err := h.DB.Save(&holder).Error; err != nil {
		jsonError(w, "An error occurred while updating the insurance holder", http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"status":  "success",
		"message": "Insurance holder updated successfully.",
	})
}

// DELETE /api/patient/insurance/{id_insurance}/holders/{patient_id}
func (h *Handler) DeleteHolder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idInsurance, _ := strconv.ParseInt(vars["id_insurance"], 10, 64)
	patientID, _ := strconv.ParseInt(vars["patient_id"], 10, 64)

	var policy insModel.InsurancePolicy
	if err := h.DB.First(&policy, idInsurance).Error; err != nil {
		jsonError(w, "Insurance policy not found", http.StatusNotFound)
		return
	}

	var holder patients.InsuranceHolderPatients
	if err := h.DB.Where("insurance_policy_id = ? AND patient_id = ?", idInsurance, patientID).
		First(&holder).Error; err != nil {
		jsonError(w, "Holder not found for this insurance policy", http.StatusNotFound)
		return
	}

	if holder.HolderType == "Self" {
		var otherCount int64
		h.DB.Model(&patients.InsuranceHolderPatients{}).
			Where("insurance_policy_id = ? AND patient_id != ?", idInsurance, patientID).
			Count(&otherCount)
		if otherCount > 0 {
			jsonError(w, "Cannot delete primary holder while other members are still associated. Remove them first.", http.StatusBadRequest)
			return
		}
		// Delete entire policy
		if err := h.DB.Delete(&policy).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, map[string]interface{}{
			"status":  "success",
			"message": "Insurance and all related holders deleted successfully.",
		})
		return
	}

	if err := h.DB.Delete(&holder).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]interface{}{
		"status":  "success",
		"message": "Holder removed from insurance policy.",
	})
}

// DELETE /api/patient/insurance/{id_insurance}/{id_patient}
func (h *Handler) DeleteInsurancePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idInsurance, _ := strconv.ParseInt(vars["id_insurance"], 10, 64)
	idPatient, _ := strconv.ParseInt(vars["id_patient"], 10, 64)

	var policy insModel.InsurancePolicy
	if err := h.DB.First(&policy, idInsurance).Error; err != nil {
		jsonError(w, "Insurance policy not found", http.StatusNotFound)
		return
	}

	var holder patients.InsuranceHolderPatients
	if err := h.DB.Where("insurance_policy_id = ? AND patient_id = ?", idInsurance, idPatient).
		First(&holder).Error; err != nil {
		jsonError(w, "Holder not found for this insurance policy", http.StatusNotFound)
		return
	}

	if holder.HolderType == "Self" {
		var otherCount int64
		h.DB.Model(&patients.InsuranceHolderPatients{}).
			Where("insurance_policy_id = ? AND holder_type != 'Self'", idInsurance).
			Count(&otherCount)
		if otherCount > 0 {
			jsonError(w, "Cannot delete primary holder while other members are still associated. Remove them first.", http.StatusBadRequest)
			return
		}
		if err := h.DB.Delete(&policy).Error; err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonOK(w, map[string]interface{}{
			"status":  "success",
			"message": "Insurance and all related holders deleted successfully.",
		})
		return
	}

	if err := h.DB.Delete(&holder).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]interface{}{
		"status":  "success",
		"message": "Holder removed from insurance policy.",
	})
}

// ─── helpers ───────────────────────────────────────────────────────────────────

func toInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	case json.Number:
		return val.Int64()
	}
	return 0, errors.New("cannot convert to int64")
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func toStrPtr(v interface{}) *string {
	s := toString(v)
	if s == "" {
		return nil
	}
	return &s
}

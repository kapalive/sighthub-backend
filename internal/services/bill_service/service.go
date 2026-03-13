package bill_service

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/marketing"
	"sighthub-backend/pkg/pdfutil"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

type FormDataResult struct {
	FormPath string                 `json:"form_path"`
	Fields   map[string]interface{} `json:"fields"`
}

// GetFormData returns path and field map for the active CMS-1500 form (id=1).
func (s *Service) GetFormData() (*FormDataResult, error) {
	var form marketing.Form
	err := s.db.Where("id_form = ? AND is_active = ?", 1, true).First(&form).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return &FormDataResult{
		FormPath: form.PathFile,
		Fields:   cms1500Fields(),
	}, nil
}

// GetFormFields returns AcroForm field IDs from the PDF of the given form.
func (s *Service) GetFormFields(formID int) ([]string, error) {
	var form marketing.Form
	if err := s.db.First(&form, formID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	if !form.IsActive {
		return nil, errors.New("not found")
	}
	fields, err := pdfutil.GetFields(form.PathFile)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.ID
	}
	return names, nil
}

// ValidateForm checks the form exists and is active (used by SubmitForm).
func (s *Service) ValidateForm(formID int) error {
	var form marketing.Form
	if err := s.db.First(&form, formID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("not found")
		}
		return err
	}
	if !form.IsActive {
		return errors.New("not found")
	}
	return nil
}

// cms1500Fields returns the CMS-1500 AcroForm field map with nil/false defaults.
func cms1500Fields() map[string]interface{} {
	n := interface{}(nil)
	f := interface{}(false)
	return map[string]interface{}{
		"insurance_name":           n,
		"insurance_address":        n,
		"insurance_address2":       n,
		"insurance_city_state_zip": n,
		"insurance_type#0":         f,
		"insurance_type#1":         f,
		"insurance_type#2":         f,
		"insurance_type#3":         f,
		"insurance_type#4":         f,
		"insurance_type#5":         f,
		"insurance_type#6":         f,
		"insurance_id":             n,
		"pt_name":                  n,
		"birth_mm":                 n,
		"birth_dd":                 n,
		"birth_yy":                 n,
		"sex#0":                    f,
		"sex#1":                    f,
		"ins_name":                 n,
		"pt_street":                n,
		"rel_to_ins#0":             f,
		"rel_to_ins#1":             f,
		"rel_to_ins#2":             f,
		"rel_to_ins#3":             f,
		"ins_street":               n,
		"pt_city":                  n,
		"pt_state":                 n,
		"ins_city":                 n,
		"ins_state":                n,
		"pt_zip":                   n,
		"pt_AreaCode":              n,
		"pt_phone":                 n,
		"NUCC USE":                 n,
		"ins_zip":                  n,
		"ins_phone area":           n,
		"ins_phone":                n,
		"other_ins_name":           n,
		"ins_policy":               n,
		"other_ins_policy":         n,
		"employment#0":             f,
		"employment#1":             f,
		"ins_sex#0":                f,
		"ins_sex#1":                f,
		"ins_dob_mm":               n,
		"ins_dob_dd":               n,
		"ins_dob_yy":               n,
		"40":                       n,
		"pt_auto_accident#0":       f,
		"other_accident#0":         f,
		"pt_auto_accident#1":       f,
		"other_accident#1":         f,
		"accident_place":           n,
		"57":                       n,
		"58":                       n,
		"41":                       n,
		"ins_plan_name":            n,
		"other_ins_plan_name":      n,
		"50":                       n,
		"pt_signature":             n,
		"pt_date":                  n,
		"ins_signature":            n,
		"cur_ill_mm":               n,
		"cur_ill_dd":               n,
		"cur_ill_yy":               n,
		"73":                       n,
		"74":                       n,
		"sim_ill_mm":               n,
		"sim_ill_dd":               n,
		"sim_ill_yy":               n,
		"work_mm_from":             n,
		"work_dd_from":             n,
		"work_yy_from":             n,
		"work_mm_end":              n,
		"work_dd_end":              n,
		"work_yy_end":              n,
		"physician number 17a1":    n,
		"physician number 17a":     n,
		"85":                       n,
		"ref_physician":            n,
		"id_physician":             n,
		"hosp_mm_from":             n,
		"hosp_dd_from":             n,
		"hosp_yy_from":             n,
		"hosp_mm_end":              n,
		"hosp_dd_end":              n,
		"hosp_yy_end":              n,
		"96":                       n,
		"ins_benefit_plan#0":       f,
		"lab#0":                    f,
		"ins_benefit_plan#1":       f,
		"lab#1":                    f,
		"charge":                   n,
		"99icd":                    n,
		"diagnosis1":               n,
		"diagnosis2":               n,
		"diagnosis3":               n,
		"diagnosis4":               n,
		"medicaid_resub":           n,
		"original_ref":             n,
		"diagnosis5":               n,
		"diagnosis6":               n,
		"diagnosis7":               n,
		"diagnosis8":               n,
		"diagnosis9":               n,
		"diagnosis10":              n,
		"diagnosis11":              n,
		"diagnosis12":              n,
		"prior_auth":               n,
		"Suppl":                    n,
		"emg1":                     n,
		"local1a":                  n,
		"sv1_mm_from":              n,
		"sv1_dd_from":              n,
		"sv1_yy_from":              n,
		"sv1_mm_end":               n,
		"sv1_dd_end":               n,
		"sv1_yy_end":               n,
		"place1":                   n,
		"type1":                    n,
		"cpt1":                     n,
		"mod1":                     n,
		"mod1a":                    n,
		"mod1b":                    n,
		"mod1c":                    n,
		"diag1":                    n,
		"ch1":                      n,
		"135":                      n,
		"day1":                     n,
		"epsdt1":                   n,
		"local1":                   n,
		"Suppla":                   n,
		"emg2":                     n,
		"local2a":                  n,
		"sv2_mm_from":              n,
		"sv2_dd_from":              n,
		"sv2_yy_from":              n,
		"sv2_mm_end":               n,
		"sv2_dd_end":               n,
		"sv2_yy_end":               n,
		"place2":                   n,
		"type2":                    n,
		"cpt2":                     n,
		"mod2":                     n,
		"mod2a":                    n,
		"mod2b":                    n,
		"mod2c":                    n,
		"diag2":                    n,
		"ch2":                      n,
		"157":                      n,
		"day2":                     n,
		"plan2":                    n,
		"local2":                   n,
		"Supplb":                   n,
		"emg3":                     n,
		"local3a":                  n,
		"sv3_mm_from":              n,
		"sv3_dd_from":              n,
		"sv3_yy_from":              n,
		"sv3_mm_end":               n,
		"sv3_dd_end":               n,
		"sv3_yy_end":               n,
		"place3":                   n,
		"type3":                    n,
		"cpt3":                     n,
		"mod3":                     n,
		"mod3a":                    n,
		"mod3b":                    n,
		"mod3c":                    n,
		"diag3":                    n,
		"ch3":                      n,
		"179":                      n,
		"day3":                     n,
		"plan3":                    n,
		"local3":                   n,
		"Supplc":                   n,
		"emg4":                     n,
		"local4a":                  n,
		"sv4_mm_from":              n,
		"sv4_dd_from":              n,
		"sv4_yy_from":              n,
		"sv4_mm_end":               n,
		"sv4_dd_end":               n,
		"sv4_yy_end":               n,
		"place4":                   n,
		"type4":                    n,
		"cpt4":                     n,
		"mod4":                     n,
		"mod4a":                    n,
		"mod4b":                    n,
		"mod4c":                    n,
		"diag4":                    n,
		"ch4":                      n,
		"201":                      n,
		"day4":                     n,
		"plan4":                    n,
		"local4":                   n,
		"Suppld":                   n,
		"emg5":                     n,
		"local5a":                  n,
		"sv5_mm_from":              n,
		"sv5_dd_from":              n,
		"sv5_yy_from":              n,
		"sv5_mm_end":               n,
		"sv5_dd_end":               n,
		"sv5_yy_end":               n,
		"place5":                   n,
		"type5":                    n,
		"cpt5":                     n,
		"mod5":                     n,
		"mod5a":                    n,
		"mod5b":                    n,
		"mod5c":                    n,
		"diag5":                    n,
		"ch5":                      n,
		"223":                      n,
		"day5":                     n,
		"plan5":                    n,
		"local5":                   n,
		"Supple":                   n,
		"emg6":                     n,
		"local6a":                  n,
		"sv6_mm_from":              n,
		"sv6_dd_from":              n,
		"sv6_yy_from":              n,
		"sv6_mm_end":               n,
		"sv6_dd_end":               n,
		"sv6_yy_end":               n,
		"place6":                   n,
		"type6":                    n,
		"cpt6":                     n,
		"mod6":                     n,
		"mod6a":                    n,
		"mod6b":                    n,
		"mod6c":                    n,
		"diag6":                    n,
		"ch6":                      n,
		"245":                      n,
		"day6":                     n,
		"plan6":                    n,
		"local6":                   n,
		"276":                      f,
		"tax_id":                   n,
		"pt_account":               n,
		"ssn#0":                    f,
		"ssn#1":                    f,
		"assignment#0":             f,
		"assignment#1":             f,
		"t_charge":                 n,
		"amt_paid":                 n,
		"doc_phone area":           n,
		"doc_phone":                n,
		"fac_name":                 n,
		"doc_name":                 n,
		"fac_street":               n,
		"doc_street":               n,
		"physician_signature":      n,
		"fac_location":             n,
		"doc_location":             n,
		"physician_date":           n,
		"pin1":                     n,
		"grp1":                     n,
		"pin":                      n,
		"grp":                      n,
		"plan1":                    n,
		"epsdt2":                   n,
		"epsdt3":                   n,
		"epsdt4":                   n,
		"epsdt5":                   n,
		"epsdt6":                   n,
	}
}

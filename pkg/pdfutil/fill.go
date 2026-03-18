// pkg/pdfutil/fill.go
// Заполнение AcroForm-полей PDF — аналог fill_cms1500 из cms1500_filler.py
package pdfutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// FillFields заполняет PDF-форму значениями из map[fieldID]value.
// Читает из inPath, пишет в outPath.
func FillFields(inPath, outPath string, values map[string]string) error {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("fill: read input: %w", err)
	}
	filled, err := FillFieldsBytes(data, values)
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, filled, 0o644)
}

// pyFillScript is an inline Python script that uses pypdf to fill AcroForm fields.
// It reads the template PDF from stdin, field values as JSON from a temp file
// (path passed as argv[1]), and writes the filled PDF to stdout.
const pyFillScript = `
import sys, json
from pypdf import PdfReader, PdfWriter
import io

fields_path = sys.argv[1]
with open(fields_path, "r") as f:
    fields = json.load(f)

reader = PdfReader(io.BytesIO(sys.stdin.buffer.read()))
writer = PdfWriter()
writer.append(reader)
writer.update_page_form_field_values(writer.pages[0], fields, auto_regenerate=True)
buf = io.BytesIO()
writer.write(buf)
sys.stdout.buffer.write(buf.getvalue())
`

// FillFieldsBytes заполняет PDF-форму и возвращает байты результата.
// Использует pypdf через subprocess (python3) для корректной работы с AcroForm.
func FillFieldsBytes(pdfData []byte, values map[string]string) (result []byte, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("fill: pypdf panic: %v", r)
		}
	}()

	// Serialize field values to a temp JSON file.
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("fill: marshal fields: %w", err)
	}
	tmpJSON, err := os.CreateTemp("", "pdfutil_fields_*.json")
	if err != nil {
		return nil, fmt.Errorf("fill: create temp json: %w", err)
	}
	defer os.Remove(tmpJSON.Name())
	if _, err := tmpJSON.Write(jsonData); err != nil {
		tmpJSON.Close()
		return nil, fmt.Errorf("fill: write temp json: %w", err)
	}
	tmpJSON.Close()

	// Run python3 with the inline script.
	cmd := exec.Command("python3", "-c", pyFillScript, tmpJSON.Name())
	cmd.Stdin = bytes.NewReader(pdfData)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("fill: pypdf failed: %w\nstderr: %s", err, stderr.String())
	}
	if stdout.Len() == 0 {
		return nil, fmt.Errorf("fill: pypdf returned empty output\nstderr: %s", stderr.String())
	}
	return stdout.Bytes(), nil
}

// ==============================================================================
// CMS-1500
// ==============================================================================

// FieldMap — маппинг DB-ключ → field_id в PDF (CMS-1500).
var FieldMap = map[string]string{
	// Тип страховки (radio): "/Medicare", "/Medicaid" и т.д.
	"insurance_type": "insurance_type",

	// Страховщик
	"insurance_company_name":           "insurance_name",
	"insurance_company_address":        "insurance_address",
	"insurance_company_address2":       "insurance_address2",
	"insurance_company_city_state_zip": "insurance_city_state_zip",
	"insurance_name":                   "insurance_name",
	"insurance_address":                "insurance_address",
	"insurance_address2":               "insurance_address2",
	"insurance_city_state_zip":         "insurance_city_state_zip",

	"insured_id_number": "insurance_id",

	"patient_name":   "pt_name",
	"patient_dob_mm": "birth_mm",
	"patient_dob_dd": "birth_dd",
	"patient_dob_yy": "birth_yy",
	"patient_sex":    "sex",

	"insured_name": "ins_name",

	"patient_address":    "pt_street",
	"patient_city":       "pt_city",
	"patient_state":      "pt_state",
	"patient_zip":        "pt_zip",
	"patient_phone_area": "pt_AreaCode",
	"patient_phone":      "pt_phone",

	"patient_relationship": "rel_to_ins",

	"insured_address":    "ins_street",
	"insured_city":       "ins_city",
	"insured_state":      "ins_state",
	"insured_zip":        "ins_zip",
	"insured_phone_area": "ins_phone area",
	"insured_phone":      "ins_phone",

	"nucc_use": "NUCC USE",

	"other_insured_name":      "other_ins_name",
	"other_insured_policy":    "other_ins_policy",
	"other_insured_plan_name": "other_ins_plan_name",

	"condition_employment":     "employment",
	"condition_auto_accident":  "pt_auto_accident",
	"accident_state":           "accident_place",
	"condition_other_accident": "other_accident",

	"insured_policy_group":     "ins_policy",
	"insured_dob_mm":           "ins_dob_mm",
	"insured_dob_dd":           "ins_dob_dd",
	"insured_dob_yy":           "ins_dob_yy",
	"insured_sex":              "ins_sex",
	"other_claim_id_qualifier": "57",
	"other_claim_id":           "58",
	"insured_plan_name":        "ins_plan_name",
	"another_benefit_plan":     "ins_benefit_plan",
	"claim_codes":              "50",

	"patient_signature":      "pt_signature",
	"patient_signature_date": "pt_date",
	"insured_signature":      "ins_signature",

	"illness_date_mm":   "cur_ill_mm",
	"illness_date_dd":   "cur_ill_dd",
	"illness_date_yy":   "cur_ill_yy",
	"illness_qualifier": "73",

	"other_date_qualifier": "74",
	"other_date_mm":        "sim_ill_mm",
	"other_date_dd":        "sim_ill_dd",
	"other_date_yy":        "sim_ill_yy",

	"work_unable_from_mm": "work_mm_from",
	"work_unable_from_dd": "work_dd_from",
	"work_unable_from_yy": "work_yy_from",
	"work_unable_to_mm":   "work_mm_end",
	"work_unable_to_dd":   "work_dd_end",
	"work_unable_to_yy":   "work_yy_end",

	"referring_provider_qualifier": "85",
	"referring_provider_name":      "ref_physician",
	"referring_provider_id_qual":   "physician number 17a1",
	"referring_provider_npi":       "physician number 17a",
	"referring_provider_other_id":  "id_physician",

	"hosp_from_mm": "hosp_mm_from",
	"hosp_from_dd": "hosp_dd_from",
	"hosp_from_yy": "hosp_yy_from",
	"hosp_to_mm":   "hosp_mm_end",
	"hosp_to_dd":   "hosp_dd_end",
	"hosp_to_yy":   "hosp_yy_end",

	"additional_claim_info": "96",
	"outside_lab":           "lab",
	"outside_lab_charges":   "charge",

	"icd_indicator": "99icd",
	"diagnosis_a":   "diagnosis1",
	"diagnosis_b":   "diagnosis2",
	"diagnosis_c":   "diagnosis3",
	"diagnosis_d":   "diagnosis4",
	"diagnosis_e":   "diagnosis5",
	"diagnosis_f":   "diagnosis6",
	"diagnosis_g":   "diagnosis7",
	"diagnosis_h":   "diagnosis8",
	"diagnosis_i":   "diagnosis9",
	"diagnosis_j":   "diagnosis10",
	"diagnosis_k":   "diagnosis11",
	"diagnosis_l":   "diagnosis12",

	"resubmission_code":   "medicaid_resub",
	"original_ref_number": "original_ref",
	"prior_auth_number":   "prior_auth",

	// Услуги строки 1–6
	"service1_from_mm": "sv1_mm_from", "service1_from_dd": "sv1_dd_from", "service1_from_yy": "sv1_yy_from",
	"service1_to_mm": "sv1_mm_end", "service1_to_dd": "sv1_dd_end", "service1_to_yy": "sv1_yy_end",
	"service1_place": "place1", "service1_emg": "emg1", "service1_cpt": "cpt1",
	"service1_modifier1": "mod1", "service1_modifier2": "mod1a", "service1_modifier3": "mod1b", "service1_modifier4": "mod1c",
	"service1_diag_pointer": "diag1", "service1_charges": "ch1", "service1_days_units": "day1",
	"service1_epsdt": "epsdt1", "service1_id_qual": "plan1", "service1_rendering_npi": "local1",
	"service1_rendering_id": "local1a", "service1_suppl": "Suppl",

	"service2_from_mm": "sv2_mm_from", "service2_from_dd": "sv2_dd_from", "service2_from_yy": "sv2_yy_from",
	"service2_to_mm": "sv2_mm_end", "service2_to_dd": "sv2_dd_end", "service2_to_yy": "sv2_yy_end",
	"service2_place": "place2", "service2_emg": "emg2", "service2_cpt": "cpt2",
	"service2_modifier1": "mod2", "service2_modifier2": "mod2a", "service2_modifier3": "mod2b", "service2_modifier4": "mod2c",
	"service2_diag_pointer": "diag2", "service2_charges": "ch2", "service2_days_units": "day2",
	"service2_epsdt": "epsdt2", "service2_id_qual": "plan2", "service2_rendering_npi": "local2",
	"service2_rendering_id": "local2a", "service2_suppl": "Suppla",

	"service3_from_mm": "sv3_mm_from", "service3_from_dd": "sv3_dd_from", "service3_from_yy": "sv3_yy_from",
	"service3_to_mm": "sv3_mm_end", "service3_to_dd": "sv3_dd_end", "service3_to_yy": "sv3_yy_end",
	"service3_place": "place3", "service3_emg": "emg3", "service3_cpt": "cpt3",
	"service3_modifier1": "mod3", "service3_modifier2": "mod3a", "service3_modifier3": "mod3b", "service3_modifier4": "mod3c",
	"service3_diag_pointer": "diag3", "service3_charges": "ch3", "service3_days_units": "day3",
	"service3_epsdt": "epsdt3", "service3_id_qual": "plan3", "service3_rendering_npi": "local3",
	"service3_rendering_id": "local3a", "service3_suppl": "Supplb",

	"service4_from_mm": "sv4_mm_from", "service4_from_dd": "sv4_dd_from", "service4_from_yy": "sv4_yy_from",
	"service4_to_mm": "sv4_mm_end", "service4_to_dd": "sv4_dd_end", "service4_to_yy": "sv4_yy_end",
	"service4_place": "place4", "service4_emg": "emg4", "service4_cpt": "cpt4",
	"service4_modifier1": "mod4", "service4_modifier2": "mod4a", "service4_modifier3": "mod4b", "service4_modifier4": "mod4c",
	"service4_diag_pointer": "diag4", "service4_charges": "ch4", "service4_days_units": "day4",
	"service4_epsdt": "epsdt4", "service4_id_qual": "plan4", "service4_rendering_npi": "local4",
	"service4_rendering_id": "local4a", "service4_suppl": "Supplc",

	"service5_from_mm": "sv5_mm_from", "service5_from_dd": "sv5_dd_from", "service5_from_yy": "sv5_yy_from",
	"service5_to_mm": "sv5_mm_end", "service5_to_dd": "sv5_dd_end", "service5_to_yy": "sv5_yy_end",
	"service5_place": "place5", "service5_emg": "emg5", "service5_cpt": "cpt5",
	"service5_modifier1": "mod5", "service5_modifier2": "mod5a", "service5_modifier3": "mod5b", "service5_modifier4": "mod5c",
	"service5_diag_pointer": "diag5", "service5_charges": "ch5", "service5_days_units": "day5",
	"service5_epsdt": "epsdt5", "service5_id_qual": "plan5", "service5_rendering_npi": "local5",
	"service5_rendering_id": "local5a", "service5_suppl": "Suppld",

	"service6_from_mm": "sv6_mm_from", "service6_from_dd": "sv6_dd_from", "service6_from_yy": "sv6_yy_from",
	"service6_to_mm": "sv6_mm_end", "service6_to_dd": "sv6_dd_end", "service6_to_yy": "sv6_yy_end",
	"service6_place": "place6", "service6_emg": "emg6", "service6_cpt": "cpt6",
	"service6_modifier1": "mod6", "service6_modifier2": "mod6a", "service6_modifier3": "mod6b", "service6_modifier4": "mod6c",
	"service6_diag_pointer": "diag6", "service6_charges": "ch6", "service6_days_units": "day6",
	"service6_epsdt": "epsdt6", "service6_id_qual": "plan6", "service6_rendering_npi": "local6",
	"service6_rendering_id": "local6a", "service6_suppl": "Supple",

	"federal_tax_id":                  "federal_tax_id",
	"tax_id_type":                     "tax_id_type",
	"patient_account_number":          "pt_acct",
	"accept_assignment":               "accept_assign",
	"total_charge":                    "total_charge",
	"amount_paid":                     "amt_paid",
	"physician_signature":             "physician_signature",
	"physician_signature_date":        "physician_signature_date",
	"facility_name":                   "facility_name",
	"facility_street":                 "facility_street",
	"facility_city_state_zip":         "facility_city_state_zip",
	"facility_npi":                    "facility_npi",
	"billing_provider_name":           "billing_provider_name",
	"billing_provider_street":         "billing_provider_street",
	"billing_provider_city_state_zip": "billing_provider_city_state_zip",
	"billing_provider_phone_area":     "billing_provider_phone_area",
	"billing_provider_phone":          "billing_provider_phone",
	"billing_provider_npi":            "billing_provider_npi",
}

var serviceSuffixes = []string{
	"from_mm", "from_dd", "from_yy", "to_mm", "to_dd", "to_yy",
	"place", "emg", "cpt",
	"modifier1", "modifier2", "modifier3", "modifier4",
	"diag_pointer", "charges", "days_units",
	"epsdt", "id_qual", "rendering_npi", "rendering_id", "suppl",
}

// FillCMS1500 заполняет CMS-1500 и сохраняет в outPath.
// Поддерживает многостраничность: 6 услуг на страницу.
func FillCMS1500(templatePath string, data map[string]string, outPath string) error {
	pages, err := prepareCMS1500Pages(templatePath, data)
	if err != nil {
		return err
	}
	if len(pages) == 1 {
		return os.WriteFile(outPath, pages[0], 0o644)
	}

	tmpFiles := make([]string, len(pages))
	for i, p := range pages {
		f, err := os.CreateTemp("", "cms1500_page_*.pdf")
		if err != nil {
			return err
		}
		_, _ = f.Write(p)
		f.Close()
		tmpFiles[i] = f.Name()
	}
	defer func() {
		for _, f := range tmpFiles {
			os.Remove(f)
		}
	}()

	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	return pdfcpuapi.MergeCreateFile(tmpFiles, outPath, false, conf)
}

// FillCMS1500Bytes — то же самое, но возвращает []byte.
func FillCMS1500Bytes(templatePath string, data map[string]string) ([]byte, error) {
	pages, err := prepareCMS1500Pages(templatePath, data)
	if err != nil {
		return nil, err
	}
	if len(pages) == 1 {
		return pages[0], nil
	}

	tmpFiles := make([]string, len(pages))
	for i, p := range pages {
		f, err := os.CreateTemp("", "cms1500_page_*.pdf")
		if err != nil {
			return nil, err
		}
		_, _ = f.Write(p)
		f.Close()
		tmpFiles[i] = f.Name()
	}
	defer func() {
		for _, f := range tmpFiles {
			os.Remove(f)
		}
	}()

	outTmp, err := os.CreateTemp("", "cms1500_merged_*.pdf")
	if err != nil {
		return nil, err
	}
	outTmp.Close()
	defer os.Remove(outTmp.Name())

	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	if err := pdfcpuapi.MergeCreateFile(tmpFiles, outTmp.Name(), false, conf); err != nil {
		return nil, err
	}
	return os.ReadFile(outTmp.Name())
}

func prepareCMS1500Pages(templatePath string, rawData map[string]string) ([][]byte, error) {
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("cms1500: read template: %w", err)
	}

	// Auto-uppercase (кроме radio-значений "/...")
	data := make(map[string]string, len(rawData))
	for k, v := range rawData {
		if v != "" && !strings.HasPrefix(v, "/") {
			data[k] = strings.ToUpper(v)
		} else {
			data[k] = v
		}
	}

	// Определяем макс. номер услуги → кол-во страниц
	maxService := 0
	for key := range data {
		if strings.HasPrefix(key, "service") {
			parts := strings.SplitN(key, "_", 2)
			n := 0
			fmt.Sscanf(strings.TrimPrefix(parts[0], "service"), "%d", &n)
			if n > maxService {
				maxService = n
			}
		}
	}
	numPages := 1
	if maxService > 6 {
		numPages = (maxService + 5) / 6
	}

	// Header-поля (всё кроме serviceN_* с N > 0)
	headerFields := make(map[string]string)
	for dbKey, value := range data {
		if strings.HasPrefix(dbKey, "service") {
			parts := strings.SplitN(dbKey, "_", 2)
			numStr := strings.TrimPrefix(parts[0], "service")
			n := 0
			fmt.Sscanf(numStr, "%d", &n)
			if n > 0 {
				continue
			}
		}
		if pdfID, ok := FieldMap[dbKey]; ok {
			headerFields[pdfID] = value
		}
	}

	pages := make([][]byte, numPages)
	for pageIdx := range numPages {
		pageFields := make(map[string]string, len(headerFields)+len(serviceSuffixes)*6)
		for k, v := range headerFields {
			pageFields[k] = v
		}
		startSvc := pageIdx*6 + 1
		for line := 1; line <= 6; line++ {
			svcNum := startSvc + line - 1
			for _, suffix := range serviceSuffixes {
				dbKey := fmt.Sprintf("service%d_%s", svcNum, suffix)
				val, ok := data[dbKey]
				if !ok {
					continue
				}
				if pdfID, ok := FieldMap[fmt.Sprintf("service%d_%s", line, suffix)]; ok {
					pageFields[pdfID] = val
				}
			}
		}

		filled, err := FillFieldsBytes(templateData, pageFields)
		if err != nil {
			return nil, fmt.Errorf("cms1500 page %d: %w", pageIdx+1, err)
		}
		pages[pageIdx] = filled
	}
	return pages, nil
}

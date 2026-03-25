package visionweb_service

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ─── VisionWeb XML structures ────────────────────────────────────────────────

type vwCatalog struct {
	XMLName      xml.Name      `xml:"CATALOG"`
	Supplier     string        `xml:"Supplier,attr"`
	SupplierName string        `xml:"SupplierName,attr"`
	Location     vwLocation    `xml:"LOCATION"`
	Lenses       []vwLens      `xml:"LENSES>LENS"`
	Treatments   []vwTreatment `xml:"TREATMENTS>TREATMENT"`
}

type vwLocation struct {
	ID   string `xml:"ID,attr"`
	Name string `xml:"Name,attr"`
}

type vwLens struct {
	RanID        string `xml:"RAN_ID,attr"`
	Description  string `xml:"Description,attr"`
	DesignCode   string `xml:"DesignCode,attr"`
	MaterialCode string `xml:"MaterialCode,attr"`
	AddID        string `xml:"ADD_ID,attr"`
}

type vwTreatment struct {
	AdtID       string `xml:"ADT_ID,attr"`
	Description string `xml:"Description,attr"`
	VwCode      string `xml:"VwCode,attr"`
}

// ─── Auth ────────────────────────────────────────────────────────────────────

const (
	tokenURL    = "https://regauth.visionweb.com/connect/token"
	catalogBase = "https://regapi.visionweb.com/catalog/catalog/Catalog"
	basicUser   = "qademo"
	basicPass   = "xNH3L7K1wl9lezUZQwdA"
	scope       = "VisionWeb.Catalog.Api"
)

func (s *Service) getToken() (string, error) {
	body := "grant_type=client_credentials&scope=" + scope
	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(basicUser, basicPass)

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("vw token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Simple JSON parse for access_token
	data, _ := io.ReadAll(resp.Body)
	s1 := strings.Index(string(data), `"access_token":"`)
	if s1 < 0 {
		return "", fmt.Errorf("vw token not found in response")
	}
	s1 += len(`"access_token":"`)
	s2 := strings.Index(string(data)[s1:], `"`)
	return string(data)[s1 : s1+s2], nil
}

func (s *Service) downloadCatalog(sloID int, token string) (*vwCatalog, error) {
	url := fmt.Sprintf("%s/%d/SUP7", catalogBase, sloID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("vw catalog download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("vw catalog HTTP %d", resp.StatusCode)
	}

	var catalog vwCatalog
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil // iso-8859-1 is a subset of UTF-8 for printable chars
	}
	if err := decoder.Decode(&catalog); err != nil {
		return nil, fmt.Errorf("vw catalog parse failed: %w", err)
	}
	return &catalog, nil
}

// ─── Material mapping ────────────────────────────────────────────────────────

var materialMap = map[string]int{
	"PL-50": 6, "PO-58": 1, "PH-55": 2, "PH-56": 2,
	"PH-60": 7, "PH-67": 19, "PH-74": 4,
	"TV-53": 9, "GL-50": 18, "GH-60": 7, "GH-70": 4, "GH-80": 4, "SP-54": 9,
}

func getMaterialID(mc string) *int {
	parts := strings.Split(mc, "-")
	if len(parts) >= 2 {
		key := parts[0] + "-" + parts[1]
		if id, ok := materialMap[key]; ok {
			return &id
		}
	}
	return nil
}

// ─── Type mapping ────────────────────────────────────────────────────────────

func getTypeID(ranID, designCode string) int {
	// DesignCode-based overrides first
	dc := strings.ToUpper(designCode)
	if strings.HasPrefix(dc, "EYEZEN") || strings.HasPrefix(dc, "RBEYEZEN") {
		return 17 // EYEZEN
	}
	if dc == "FT35" || dc == "FT45" || dc == "DOUBLEST35" || dc == "FT835" {
		return 7 // BIFOCAL FT-35
	}
	if strings.HasPrefix(dc, "RD22") || strings.HasPrefix(dc, "DOUBLE") && strings.Contains(dc, "RD22") {
		return 6 // ROUNDSEG R22
	}
	if strings.HasPrefix(dc, "RD24") || strings.HasPrefix(dc, "RD25") {
		return 9 // ROUNDSEG R24
	}
	// SV-based
	if dc == "SV" || dc == "ASV" || strings.HasSuffix(dc, "SV") || strings.HasSuffix(dc, "SVH") {
		return 2 // SINGLE
	}

	// RAN_ID-based
	if strings.HasPrefix(ranID, "SV") {
		return 2
	}
	if strings.HasPrefix(ranID, "PAL") {
		return 5 // PROGRESSIVE
	}
	if strings.HasPrefix(ranID, "BF") {
		return 3 // BIFOCAL FT-28
	}
	if strings.HasPrefix(ranID, "TF") {
		return 4 // TRIFOCAL
	}
	return 11 // OTHER
}

// ─── Material name for description ───────────────────────────────────────────

var matNameMap = map[string]string{
	"PL-50": "PLASTIC", "PO-58": "POLYCARBONATE", "PH-55": "MID INDEX 1.56",
	"PH-56": "MID INDEX 1.56", "PH-60": "HIGH INDEX 1.60", "PH-67": "PLASTIC 1.67",
	"PH-74": "HIGH INDEX 1.74", "TV-53": "TRIVEX", "GL-50": "GLASS",
	"GH-60": "GLASS HI 1.60", "GH-70": "GLASS HI 1.70", "GH-80": "GLASS HI 1.80",
}

var coatNameMap = map[string]string{
	"NONE-NONE": "", "NONE-BLGD": "BLUE GUARD",
	"PHFX-GRY0": "PHOTOFUSION GRAY", "PHFX-BRN0": "PHOTOFUSION BROWN",
	"PHFX-GRYX": "PHOTOFUSION XTRA GRAY", "PHFX-BLUX": "PHOTOFUSION XTRA BLUE",
	"PHFX-GRNX": "PHOTOFUSION XTRA GREEN", "PHFX-BURX": "PHOTOFUSION XTRA BURGUNDY",
	"PHXP-GRY0": "PHOTOFUSION XTRA POLAR GRAY", "PHXP-BRN0": "PHOTOFUSION XTRA POLAR BROWN",
	"POLR-GRY3": "POLARIZED GRAY", "POLR-BRN3": "POLARIZED BROWN",
	"POLR-SKLF": "POLARIZED SKYLET FUN", "POLR-SKLR": "POLARIZED SKYLET ROAD",
	"POLR-SKLS": "POLARIZED SKYLET SPORT", "POLR-GRNX": "POLARIZED GREEN",
	"POLR-G150": "POLARIZED GRAY 15",
	"TRAN-GRY9": "TRANSITIONS GRAY", "TRAN-BRN9": "TRANSITIONS BROWN",
	"TRAN-GG09": "TRANSITIONS GEN8 GRAY", "TRAN-RBY9": "TRANSITIONS RUBY",
	"TRAN-AMB9": "TRANSITIONS AMBER", "TRAN-AMT9": "TRANSITIONS AMETHYST",
	"TRAN-EMR9": "TRANSITIONS EMERALD", "TRAN-SPH9": "TRANSITIONS SAPPHIRE",
	"TRAN-TVG0": "TRANSITIONS TGNS GRAY",
	"XTRP-GRY0": "XTRACTIVE POLARIZED GRAY", "XTRA-GRY0": "XTRACTIVE GRAY",
	"XTNG-GRY0": "XTRACTIVE NEW GEN GRAY", "XTNG-BRN0": "XTRACTIVE NEW GEN BROWN",
	"XTNG-GGR0": "XTRACTIVE NEW GEN GRADIENT GRAY",
	"SNSC-GRY0": "SENSITY GRAY", "SNSC-BRN0": "SENSITY BROWN",
	"SNSE-GRY0": "SENSITY EXTRA GRAY", "SNSE-BRN0": "SENSITY EXTRA BROWN",
	"SSXT-GRY0": "SENSITY XTRACTIVE GRAY", "SSXT-BRN0": "SENSITY XTRACTIVE BROWN",
	"EEBS-NONE": "EYE PROTECT SYSTEM", "BLUZ-NONE": "BLUE UV FILTER",
	"ECT0-NONE": "EYECONNECT", "ECTP-NONE": "EYECONNECT PLUS",
	"PHPO-DWR0": "PHOTO POLARIZED",
}

func buildLensDescription(brand, design, typeName, matCode string) string {
	parts := strings.Split(matCode, "-")
	matName := ""
	if len(parts) >= 2 {
		key := parts[0] + "-" + parts[1]
		matName = matNameMap[key]
		if matName == "" {
			matName = key
		}
	}
	coat := ""
	if len(parts) >= 4 {
		key := parts[2] + "-" + parts[3]
		coat = coatNameMap[key]
	}
	result := strings.Join(filterEmpty([]string{brand, design, typeName, matName, coat}), " ")
	return result
}

func filterEmpty(ss []string) []string {
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ─── Special features from MaterialCode ──────────────────────────────────────

func getLensSpecialFeatures(matCode string) []int {
	parts := strings.Split(matCode, "-")
	var feats []int

	if len(parts) >= 3 {
		suffix := strings.ToUpper(strings.Join(parts[2:], "-"))
		photoKeys := []string{"PHFX", "PHXP", "TRAN", "XTRA", "XTNG", "XTRP", "SNSC", "SNSE", "SSXT", "PHPO"}
		for _, k := range photoKeys {
			if strings.Contains(suffix, k) {
				feats = append(feats, 1) // PHOTOCHROMATIC
				break
			}
		}
		if strings.Contains(suffix, "POLR") {
			feats = append(feats, 2) // POLARIZED
		}
		if strings.Contains(suffix, "BLGD") || strings.Contains(suffix, "BLUZ") || strings.Contains(suffix, "EEBS") {
			feats = append(feats, 5) // DIGITAL
		}
	}

	// HIGH INDEX from material
	if len(parts) >= 2 {
		prefix := parts[0]
		idx := parts[1]
		if prefix == "PH" && (idx == "60" || idx == "67" || idx == "74") {
			feats = append(feats, 4) // HIGH INDEX
		}
	}

	return feats
}

func getTreatmentSpecialFeatures(vwCode, desc string) []int {
	uc := strings.ToUpper(vwCode + " " + desc)
	var feats []int
	if strings.Contains(uc, "AR") || strings.Contains(uc, "COAT") || strings.Contains(uc, "CRIZAL") || strings.Contains(uc, "DURAVIS") {
		feats = append(feats, 9) // AR COATING
	}
	if strings.Contains(uc, "TINT") {
		feats = append(feats, 10) // TINT
	}
	if strings.Contains(uc, "SCRATCH") || strings.HasPrefix(strings.ToUpper(vwCode), "SR-") {
		feats = append(feats, 12) // SCRATCH RESISTANT
	}
	return feats
}

// ─── V Code mapping ──────────────────────────────────────────────────────────

type typeMatKey struct{ typeID, matID int }

var vCodeMap = map[typeMatKey]int{
	{2, 6}: 1, {2, 1}: 1, {2, 9}: 1, {2, 7}: 1, {2, 19}: 1, {2, 4}: 1, {2, 2}: 1,
	{2, 18}: 2,
	{3, 6}: 3, {3, 1}: 3, {3, 9}: 3, {3, 7}: 3, {3, 19}: 3, {3, 4}: 3, {3, 18}: 3,
	{7, 6}: 3, {7, 1}: 3,
	{4, 6}: 4, {4, 1}: 4, {4, 9}: 4, {4, 18}: 4,
	{5, 6}: 11, {5, 1}: 6, {5, 9}: 6, {5, 7}: 7, {5, 19}: 8, {5, 4}: 7, {5, 2}: 11, {5, 18}: 5,
	{17, 6}: 1, {17, 1}: 1, {17, 9}: 1, {17, 7}: 1, {17, 19}: 1, {17, 4}: 1,
	{6, 6}: 3, {6, 1}: 3, {9, 6}: 3, {9, 1}: 3,
}

// ─── Series mapping ──────────────────────────────────────────────────────────

func (s *Service) getSeriesMap() map[string]int {
	type row struct {
		ID   int
		Name string
	}
	var rows []row
	s.db.Table("lens_series").Select("id_lens_series AS id, UPPER(series_name) AS name").Scan(&rows)
	m := make(map[string]int, len(rows))
	for _, r := range rows {
		m[r.Name] = r.ID
	}
	return m
}

// ─── Import result ───────────────────────────────────────────────────────────

type ImportResult struct {
	SupplierName      string `json:"supplier_name"`
	LensesImported    int    `json:"lenses_imported"`
	LensesSkipped     int    `json:"lenses_skipped"`
	TreatmentsImported int   `json:"treatments_imported"`
	TreatmentsSkipped  int   `json:"treatments_skipped"`
}

// ─── Import All VW Labs ──────────────────────────────────────────────────────

type ImportAllResult struct {
	Labs []ImportResult `json:"labs"`
}

func (s *Service) ImportAllVisionWeb() (*ImportAllResult, error) {
	// Get all vendor_location_accounts with VW slo_id configured
	type vlaRow struct {
		VendorID    int    `gorm:"column:vendor_id"`
		VwSloID     int    `gorm:"column:vw_slo_id"`
		VendorName  string `gorm:"column:vendor_name"`
	}
	var vlas []vlaRow
	s.db.Table("vendor_location_account vla").
		Select("DISTINCT vla.vendor_id, vla.vw_slo_id, v.vendor_name").
		Joins("JOIN vendor v ON v.id_vendor = vla.vendor_id").
		Where("vla.source = 'vision_web' AND vla.vw_slo_id IS NOT NULL").
		Scan(&vlas)

	if len(vlas) == 0 {
		return nil, fmt.Errorf("no labs configured for VisionWeb import")
	}

	result := &ImportAllResult{}
	for _, vla := range vlas {
		// Find brand_lens for this vendor
		var brandLensID int
		s.db.Table("brand_lens bl").
			Select("bl.id_brand_lens").
			Joins("JOIN lenses l ON l.brand_lens_id = bl.id_brand_lens").
			Where("l.vendor_id = ? AND l.source = 'vision_web'", vla.VendorID).
			Limit(1).Scan(&brandLensID)
		if brandLensID == 0 {
			// No existing brand — try lab table
			s.db.Table("lab").Select("brand_lens_id").Where("vendor_id = ?", vla.VendorID).Scan(&brandLensID)
		}
		if brandLensID == 0 {
			// Create brand from vendor name
			s.db.Exec("INSERT INTO brand_lens (brand_name) VALUES (?) ON CONFLICT DO NOTHING", vla.VendorName)
			s.db.Table("brand_lens").Select("id_brand_lens").Where("brand_name = ?", vla.VendorName).Scan(&brandLensID)
		}

		r, err := s.ImportFromVisionWeb(vla.VwSloID, vla.VendorID, brandLensID)
		if err != nil {
			result.Labs = append(result.Labs, ImportResult{
				SupplierName: vla.VendorName + " (error: " + err.Error() + ")",
			})
			continue
		}
		result.Labs = append(result.Labs, *r)
	}
	return result, nil
}

// ─── Import single lab ───────────────────────────────────────────────────────

func (s *Service) ImportFromVisionWeb(sloID, vendorID, brandLensID int) (*ImportResult, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	catalog, err := s.downloadCatalog(sloID, token)
	if err != nil {
		return nil, err
	}

	// Get type names for description
	typeNames := map[int]string{}
	type tnRow struct {
		ID   int
		Name string
	}
	var tnRows []tnRow
	s.db.Table("lens_types").Select("id_lens_type AS id, type_name AS name").Scan(&tnRows)
	for _, r := range tnRows {
		typeNames[r.ID] = r.Name
	}

	// Get vendor name
	var vendorName string
	s.db.Table("vendor").Select("vendor_name").Where("id_vendor = ?", vendorID).Scan(&vendorName)

	// Get brand name
	var brandName string
	s.db.Table("brand_lens").Select("brand_name").Where("id_brand_lens = ?", brandLensID).Scan(&brandName)

	seriesMap := s.getSeriesMap()
	genericSeriesID := seriesMap["GENERIC"]

	// Get next L-code for lenses
	var maxLCode int
	s.db.Raw("SELECT COALESCE(MAX(CAST(SUBSTRING(lens_name FROM 2) AS INTEGER)), 7000) FROM lenses WHERE lens_name ~ '^L[0-9]+$'").Scan(&maxLCode)
	nextLCode := maxLCode + 1

	// Get next A-code for treatments
	var maxACode int
	s.db.Raw("SELECT COALESCE(MAX(CAST(SUBSTRING(item_nbr FROM 2) AS INTEGER)), 7000) FROM lens_treatments WHERE item_nbr ~ '^A[0-9]+$'").Scan(&maxACode)
	nextACode := maxACode + 1

	result := &ImportResult{SupplierName: catalog.SupplierName}

	// ─── Import Lenses ───────────────────────────────────────────────────
	for _, l := range catalog.Lenses {
		// Check if already exists
		var existingID int
		s.db.Table("lenses").Select("id_lenses").
			Where("vw_design_code = ? AND vw_material_code = ? AND vendor_id = ?",
				l.DesignCode, l.MaterialCode, vendorID).Scan(&existingID)
		if existingID > 0 {
			result.LensesSkipped++
			continue
		}

		matID := getMaterialID(l.MaterialCode)
		typeID := getTypeID(l.RanID, l.DesignCode)
		typeName := typeNames[typeID]
		desc := buildLensDescription(brandName, l.DesignCode, typeName, l.MaterialCode)
		seriesID := genericSeriesID

		lensName := fmt.Sprintf("L%d", nextLCode)
		nextLCode++

		src := "vision_web"
		s.db.Exec(`
			INSERT INTO lenses (lens_name, description, vendor_id, brand_lens_id,
				lens_type_id, lenses_materials_id, lens_series_id, source, can_lookup,
				vw_design_code, vw_material_code, vw_add_id, vw_ran_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, true, ?, ?, ?, ?)`,
			lensName, desc, vendorID, brandLensID,
			typeID, matID, seriesID, src,
			l.DesignCode, l.MaterialCode, l.AddID, l.RanID)

		// Get inserted ID
		var lensID int
		s.db.Table("lenses").Select("id_lenses").
			Where("vw_design_code = ? AND vw_material_code = ? AND vendor_id = ?",
				l.DesignCode, l.MaterialCode, vendorID).Scan(&lensID)

		if lensID > 0 {
			// Special features
			for _, sfID := range getLensSpecialFeatures(l.MaterialCode) {
				s.db.Exec("INSERT INTO lenses_feature_relation (lenses_id, lens_special_features_id) VALUES (?, ?) ON CONFLICT DO NOTHING", lensID, sfID)
			}
			// V codes
			if matID != nil {
				vcID, ok := vCodeMap[typeMatKey{typeID, *matID}]
				if !ok {
					vcID = 10 // V2799
				}
				s.db.Exec("INSERT INTO lenses_v_codes_relation (lenses_id, v_codes_lens_id) VALUES (?, ?) ON CONFLICT DO NOTHING", lensID, vcID)
			}
		}

		result.LensesImported++
	}

	// ─── Import Treatments ───────────────────────────────────────────────
	for _, t := range catalog.Treatments {
		// Check duplicate
		var exists int64
		s.db.Table("lens_treatments").Where("vw_code = ? AND vendor_id = ?", t.VwCode, vendorID).Count(&exists)
		if exists > 0 {
			result.TreatmentsSkipped++
			continue
		}

		itemNbr := fmt.Sprintf("A%d", nextACode)
		nextACode++
		desc := t.Description

		s.db.Exec(`
			INSERT INTO lens_treatments (item_nbr, description, vendor_id, source, can_lookup, vw_code, vw_adt_id)
			VALUES (?, ?, ?, 'vision_web', true, ?, ?)`,
			itemNbr, desc, vendorID, t.VwCode, t.AdtID)

		// Get inserted ID
		var treatID int64
		s.db.Table("lens_treatments").Select("id_lens_treatments").
			Where("vw_code = ? AND vendor_id = ?", t.VwCode, vendorID).Scan(&treatID)

		if treatID > 0 {
			for _, sfID := range getTreatmentSpecialFeatures(t.VwCode, t.Description) {
				s.db.Exec("INSERT INTO treatments_feature_relation (lens_treatments_id, lens_special_features_id) VALUES (?, ?) ON CONFLICT DO NOTHING", treatID, sfID)
			}
		}

		result.TreatmentsImported++
	}

	return result, nil
}

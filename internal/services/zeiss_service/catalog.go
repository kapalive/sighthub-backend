package zeiss_service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	lensmodel "sighthub-backend/internal/models/lenses"
	vendormodel "sighthub-backend/internal/models/vendors"

	"gorm.io/gorm"
)

// ─── Constants ───────────────────────────────────────────────────────────────

const ZeissVendorID = 69 // CARL ZEISS vendor
const zeissSource = "zeiss_only"

// ─── PCAT API response structs ──────────────────────────────────────────────

type pcatLensInfo struct {
	Code             string `json:"code"`
	ID               int    `json:"id"`
	Name             string `json:"name"`
	BrandName        string `json:"brandName"`
	FocalType        string `json:"focalType"`
	Index            string `json:"index"`
	MaterialType     string `json:"materialType"`
	MaterialProperty string `json:"materialProperty"`
	MaterialVariant  *struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"materialVariant"`
	ProductFamily *struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"productFamily"`
	ProductLines []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"productLines"`
	Category *struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	Stock bool `json:"stock"`
}

type pcatTreatmentOption struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Type          string  `json:"type"` // COATING or COLOR
	NecessityType string  `json:"necessityType"`
	Sort          float64 `json:"sort"`
}

type pcatLensTreatments struct {
	Code    string                `json:"code"`
	ID      int                   `json:"id"`
	Name    string                `json:"name"`
	Options []pcatTreatmentOption `json:"options"`
}

// ─── ImportResult ────────────────────────────────────────────────────────────

type ImportResult struct {
	LensesCreated     int      `json:"lenses_created"`
	LensesUpdated     int      `json:"lenses_updated"`
	BrandsCreated     int      `json:"brands_created"`
	MaterialsCreated  int      `json:"materials_created"`
	TreatmentsCreated int      `json:"treatments_created"`
	TreatmentsUpdated int      `json:"treatments_updated"`
	Errors            []string `json:"errors,omitempty"`
}

// ─── CatalogService ─────────────────────────────────────────────────────────

type CatalogService struct {
	db   *gorm.DB
	auth *AuthService
}

func NewCatalogService(db *gorm.DB, auth *AuthService) *CatalogService {
	return &CatalogService{db: db, auth: auth}
}

// ─── ImportCatalog ──────────────────────────────────────────────────────────

func (s *CatalogService) ImportCatalog(employeeID int64, customerNumber string) (*ImportResult, error) {
	result := &ImportResult{}

	// 1. Get access token
	token, err := s.auth.GetToken(employeeID)
	if err != nil {
		return nil, fmt.Errorf("zeiss catalog: %w", err)
	}

	// 2. Fetch generic-information
	path := fmt.Sprintf("/public/api/catalogue/v1/lenses/generic-information/%s", customerNumber)
	body, err := s.pcatGET(context.Background(), token, path)
	if err != nil {
		return nil, fmt.Errorf("zeiss catalog: fetch lenses: %w", err)
	}

	// 3. Parse response
	var lenses []pcatLensInfo
	if err := json.Unmarshal(body, &lenses); err != nil {
		return nil, fmt.Errorf("zeiss catalog: parse lenses: %w", err)
	}

	log.Printf("[zeiss] fetched %d lenses for customer %s", len(lenses), customerNumber)

	// Local caches to avoid repeated DB queries
	brandCache := make(map[string]int)       // brand_name → id_brand_lens
	lensTypeCache := make(map[string]*int)   // focalType → id_lens_type (nil = not found)
	materialCache := make(map[string]int)     // cacheKey → id_lenses_materials
	vblCache := make(map[string]bool)         // "vendorID:brandID" → exists

	// Collect all PCAT lens IDs for treatment fetch later
	pcatLensIDs := make([]int, 0, len(lenses))

	// 4. Process each lens
	for _, lens := range lenses {
		if err := s.processLens(lens, token, result, brandCache, lensTypeCache, materialCache, vblCache); err != nil {
			errMsg := fmt.Sprintf("lens %s: %v", lens.Code, err)
			log.Printf("[zeiss] error: %s", errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}
		pcatLensIDs = append(pcatLensIDs, lens.ID)
	}

	// 5. Fetch and import treatments (batched by 50)
	s.importTreatments(token, customerNumber, pcatLensIDs, result)

	log.Printf("[zeiss] import done: lenses created=%d updated=%d, brands=%d, materials=%d, treatments created=%d updated=%d, errors=%d",
		result.LensesCreated, result.LensesUpdated, result.BrandsCreated,
		result.MaterialsCreated, result.TreatmentsCreated, result.TreatmentsUpdated,
		len(result.Errors))

	return result, nil
}

// ─── processLens ────────────────────────────────────────────────────────────

func (s *CatalogService) processLens(
	lens pcatLensInfo,
	token string,
	result *ImportResult,
	brandCache map[string]int,
	lensTypeCache map[string]*int,
	materialCache map[string]int,
	vblCache map[string]bool,
) error {
	// a. Determine brand name
	brandName := s.determineBrandName(lens)

	// b. Find or create brand_lens
	brandID, created, err := s.findOrCreateBrand(brandName, brandCache)
	if err != nil {
		return fmt.Errorf("brand: %w", err)
	}
	if created {
		result.BrandsCreated++
	}

	// c. Ensure vendor_brand_lens link
	vblKey := fmt.Sprintf("%d:%d", ZeissVendorID, brandID)
	if !vblCache[vblKey] {
		if err := s.ensureVendorBrandLens(ZeissVendorID, brandID); err != nil {
			return fmt.Errorf("vendor_brand_lens: %w", err)
		}
		vblCache[vblKey] = true
	}

	// d. Map focalType to lens_type
	lensTypeID, err := s.findLensType(lens.FocalType, lensTypeCache)
	if err != nil {
		return fmt.Errorf("lens_type: %w", err)
	}

	// e. Map index + materialType to lenses_materials
	materialID, created, err := s.findOrCreateMaterial(lens.Index, lens.MaterialType, materialCache)
	if err != nil {
		return fmt.Errorf("material: %w", err)
	}
	if created {
		result.MaterialsCreated++
	}

	// f. Build description: inject focalType if not already in name
	desc := buildZeissDescription(lens.Name, lens.FocalType)

	// Upsert lens
	vendorID := ZeissVendorID
	source := zeissSource
	var existing lensmodel.Lenses
	err = s.db.Where("lens_name = ? AND vendor_id = ? AND source = ?", lens.Code, ZeissVendorID, zeissSource).
		First(&existing).Error

	if err == nil {
		// Update existing
		updates := map[string]interface{}{
			"description":        desc,
			"brand_lens_id":      brandID,
			"lenses_materials_id": materialID,
		}
		if lensTypeID != nil {
			updates["lens_type_id"] = *lensTypeID
		}
		if err := s.db.Model(&existing).Updates(updates).Error; err != nil {
			return fmt.Errorf("update lens: %w", err)
		}
		result.LensesUpdated++
	} else if err == gorm.ErrRecordNotFound {
		// Create new
		zero := float64(0)
		code := lens.Code
		newLens := lensmodel.Lenses{
			LensName:          lens.Code,
			Description:       &desc,
			VendorID:          &vendorID,
			BrandLensID:       &brandID,
			LensesMaterialsID: &materialID,
			Source:             &source,
			CanLookup:         true,
			Price:             &zero,
			Cost:              &zero,
			VwDesignCode:      &code,
		}
		if lensTypeID != nil {
			newLens.LensTypeID = lensTypeID
		}
		if err := s.db.Create(&newLens).Error; err != nil {
			return fmt.Errorf("create lens: %w", err)
		}
		result.LensesCreated++
	} else {
		return fmt.Errorf("query lens: %w", err)
	}

	return nil
}

// ─── buildZeissDescription ──────────────────────────────────────────────────

var focalTypeLabel = map[string]string{
	"PROGRESSIVE":  "Progressive",
	"SINGLEVISION": "Single Vision",
	"BIFOCAL":      "Bifocal",
	"TRIFOCAL":     "Trifocal",
}

func buildZeissDescription(name, focalType string) string {
	label, ok := focalTypeLabel[focalType]
	if !ok || label == "" {
		return name
	}
	// If name already contains the focal type keyword, don't duplicate
	nameLower := strings.ToLower(name)
	if strings.Contains(nameLower, strings.ToLower(label)) {
		return name
	}
	// Also check partial matches: "Single" for "Single Vision", "Digital" covers progressive
	if focalType == "SINGLEVISION" && strings.Contains(nameLower, "single") {
		return name
	}
	// Insert focal type after brand prefix (first word or "ZEISS")
	// e.g. "Ethos 1.5" → "Ethos Progressive 1.5"
	// e.g. "ZEISS Digital Lens SmartLife..." → already has Digital, but focalType=PROGRESSIVE
	// For progressive, "Digital Lens" implies progressive — skip
	if focalType == "PROGRESSIVE" && strings.Contains(nameLower, "digital") {
		return name
	}
	// Insert after first word: "Ethos 1.5" → "Ethos Progressive 1.5"
	parts := strings.SplitN(name, " ", 2)
	if len(parts) == 2 {
		return parts[0] + " " + label + " " + parts[1]
	}
	return name + " " + label
}

// ─── determineBrandName ─────────────────────────────────────────────────────

func (s *CatalogService) determineBrandName(lens pcatLensInfo) string {
	if lens.ProductFamily != nil && lens.ProductFamily.Name != "" {
		if len(lens.ProductLines) > 0 && lens.ProductLines[0].Name != "" {
			return lens.ProductLines[0].Name + " " + lens.ProductFamily.Name
		}
		return lens.ProductFamily.Name
	}
	if lens.BrandName != "" {
		return lens.BrandName
	}
	return "ZEISS"
}

// ─── findOrCreateBrand ──────────────────────────────────────────────────────

func (s *CatalogService) findOrCreateBrand(brandName string, cache map[string]int) (int, bool, error) {
	if id, ok := cache[brandName]; ok {
		return id, false, nil
	}

	var brand vendormodel.BrandLens
	err := s.db.Where("brand_name = ?", brandName).First(&brand).Error
	if err == nil {
		cache[brandName] = brand.IDBrandLens
		return brand.IDBrandLens, false, nil
	}
	if err != gorm.ErrRecordNotFound {
		return 0, false, err
	}

	// Create new brand
	brand = vendormodel.BrandLens{
		BrandName: brandName,
		CanLookup: true,
	}
	if err := s.db.Create(&brand).Error; err != nil {
		return 0, false, fmt.Errorf("create brand_lens: %w", err)
	}
	cache[brandName] = brand.IDBrandLens
	return brand.IDBrandLens, true, nil
}

// ─── ensureVendorBrandLens ──────────────────────────────────────────────────

func (s *CatalogService) ensureVendorBrandLens(vendorID, brandLensID int) error {
	var count int64
	if err := s.db.Model(&vendormodel.VendorBrandLens{}).
		Where("id_vendor = ? AND id_brand_lens = ?", vendorID, brandLensID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	vbl := vendormodel.VendorBrandLens{
		IDVendor:    vendorID,
		IDBrandLens: brandLensID,
	}
	return s.db.Create(&vbl).Error
}

// ─── findLensType ───────────────────────────────────────────────────────────

func (s *CatalogService) findLensType(focalType string, cache map[string]*int) (*int, error) {
	if val, ok := cache[focalType]; ok {
		return val, nil
	}

	var pattern string
	switch strings.ToUpper(focalType) {
	case "PROGRESSIVE":
		pattern = "%progressive%"
	case "SINGLEVISION":
		pattern = "%single%"
	case "BIFOCAL":
		pattern = "%bifocal%"
	case "TRIFOCAL":
		pattern = "%trifocal%"
	default:
		cache[focalType] = nil
		return nil, nil
	}

	var lt lensmodel.LensType
	err := s.db.Where("LOWER(type_name) LIKE ?", pattern).First(&lt).Error
	if err == gorm.ErrRecordNotFound {
		cache[focalType] = nil
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	id := lt.IDLensType
	cache[focalType] = &id
	return &id, nil
}

// ─── findOrCreateMaterial ───────────────────────────────────────────────────

func (s *CatalogService) findOrCreateMaterial(index, materialType string, cache map[string]int) (int, bool, error) {
	cacheKey := index + "|" + materialType
	if id, ok := cache[cacheKey]; ok {
		return id, false, nil
	}

	// Try to find by index value
	var mat lensmodel.LensesMaterial
	indexVal, parseErr := strconv.ParseFloat(index, 64)

	if parseErr == nil {
		err := s.db.Where("index = ?", indexVal).First(&mat).Error
		if err == nil {
			cache[cacheKey] = int(mat.IDLensesMaterials)
			return int(mat.IDLensesMaterials), false, nil
		}
		if err != gorm.ErrRecordNotFound {
			return 0, false, err
		}
	}

	// Try to find by material_name pattern
	searchName := index + " " + materialType
	err := s.db.Where("LOWER(material_name) LIKE ?", "%"+strings.ToLower(searchName)+"%").
		First(&mat).Error
	if err == nil {
		cache[cacheKey] = int(mat.IDLensesMaterials)
		return int(mat.IDLensesMaterials), false, nil
	}
	if err != gorm.ErrRecordNotFound {
		return 0, false, err
	}

	// Create new material
	matName := "Index " + index
	mat = lensmodel.LensesMaterial{
		MaterialName: matName,
	}
	if parseErr == nil {
		mat.Index = &indexVal
	}
	if err := s.db.Create(&mat).Error; err != nil {
		return 0, false, fmt.Errorf("create lenses_materials: %w", err)
	}
	cache[cacheKey] = int(mat.IDLensesMaterials)
	return int(mat.IDLensesMaterials), true, nil
}

// ─── importTreatments ───────────────────────────────────────────────────────

func (s *CatalogService) importTreatments(token, customerNumber string, pcatLensIDs []int, result *ImportResult) {
	if len(pcatLensIDs) == 0 {
		return
	}

	// Collect all unique treatments across batches
	seen := make(map[string]pcatTreatmentOption) // code → option

	// Batch by 50
	for i := 0; i < len(pcatLensIDs); i += 50 {
		end := i + 50
		if end > len(pcatLensIDs) {
			end = len(pcatLensIDs)
		}
		batch := pcatLensIDs[i:end]

		// Build comma-separated IDs
		ids := make([]string, len(batch))
		for j, id := range batch {
			ids[j] = strconv.Itoa(id)
		}
		idsParam := strings.Join(ids, ",")

		path := fmt.Sprintf("/public/api/catalogue/v1/lenses/treatment-options/%s?lensids=%s",
			customerNumber, idsParam)
		body, err := s.pcatGET(context.Background(), token, path)
		if err != nil {
			errMsg := fmt.Sprintf("fetch treatments batch %d-%d: %v", i, end, err)
			log.Printf("[zeiss] error: %s", errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}

		var treatments []pcatLensTreatments
		if err := json.Unmarshal(body, &treatments); err != nil {
			errMsg := fmt.Sprintf("parse treatments batch %d-%d: %v", i, end, err)
			log.Printf("[zeiss] error: %s", errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}

		for _, t := range treatments {
			for _, opt := range t.Options {
				if _, exists := seen[opt.Code]; !exists {
					seen[opt.Code] = opt
				}
			}
		}
	}

	log.Printf("[zeiss] collected %d unique treatments", len(seen))

	// Get next A-code
	var maxACode int
	s.db.Raw("SELECT COALESCE(MAX(CAST(SUBSTRING(item_nbr FROM 2) AS INTEGER)), 7000) FROM lens_treatments WHERE item_nbr ~ '^A[0-9]+$'").Scan(&maxACode)
	nextACode := maxACode + 1

	// Upsert treatments (match by vw_code + vendor_id + source)
	source := zeissSource
	for code, opt := range seen {
		var existing lensmodel.LensTreatments
		err := s.db.Where("vw_code = ? AND vendor_id = ? AND source = ?", code, ZeissVendorID, zeissSource).
			First(&existing).Error

		if err == nil {
			// Update description
			if err := s.db.Model(&existing).Update("description", opt.Name).Error; err != nil {
				errMsg := fmt.Sprintf("update treatment %s: %v", code, err)
				log.Printf("[zeiss] error: %s", errMsg)
				result.Errors = append(result.Errors, errMsg)
				continue
			}
			result.TreatmentsUpdated++
		} else if err == gorm.ErrRecordNotFound {
			zero := float64(0)
			desc := opt.Name
			itemNbr := fmt.Sprintf("A%d", nextACode)
			nextACode++
			vwCode := code
			treatment := lensmodel.LensTreatments{
				ItemNbr:     itemNbr,
				Description: &desc,
				Price:       &zero,
				Cost:        &zero,
				VendorID:    ZeissVendorID,
				Source:      &source,
				VwCode:      &vwCode,
				CanLookup:   true,
			}
			if err := s.db.Create(&treatment).Error; err != nil {
				errMsg := fmt.Sprintf("create treatment %s: %v", code, err)
				log.Printf("[zeiss] error: %s", errMsg)
				result.Errors = append(result.Errors, errMsg)
				continue
			}
			result.TreatmentsCreated++
		} else {
			errMsg := fmt.Sprintf("query treatment %s: %v", code, err)
			log.Printf("[zeiss] error: %s", errMsg)
			result.Errors = append(result.Errors, errMsg)
		}
	}
}

// ─── pcatGET ────────────────────────────────────────────────────────────────

func (s *CatalogService) pcatGET(ctx context.Context, token string, path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	url := s.auth.APIBase() + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", s.auth.PCatSubKey())
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

package price_book_service

import (
	"fmt"

	"sighthub-backend/internal/models/service"
)

// ─── Result types ─────────────────────────────────────────────────────────────

type ProfServiceTypeResult struct {
	TypeID int    `json:"type_id"`
	Title  string `json:"title"`
}

type ProfServiceScopeResult struct {
	ScopeID int    `json:"scope_id"`
	Title   string `json:"title"`
}

type ProfServiceListItem struct {
	ItemID       int64   `json:"item_id"`
	ItemName     string  `json:"item_name"`
	CptHcpcsCode *string `json:"cpt_hcpcs_code"`
	Scope        *string `json:"scope"`
	Type         *string `json:"type"`
	Description  *string `json:"description"`
	Price        *string `json:"price"`
	Cost         *string `json:"cost"`
	PbKey        string  `json:"pb_key"`
}

type ProfServiceDetail struct {
	ServiceID           int64                  `json:"service_id"`
	ItemNumber          string                 `json:"item_number"`
	CptHcpcsCode        *string                `json:"cpt_hcpcs_code"`
	Scope               map[string]interface{} `json:"scope"`
	Type                map[string]interface{} `json:"type"`
	Description         *string                `json:"description"`
	Price               *string                `json:"price"`
	Cost                *string                `json:"cost"`
	Sort1               *float64               `json:"sort1"`
	Sort2               *float64               `json:"sort2"`
	ReferringPhysician  bool                   `json:"referring_physician"`
	Visible             bool                   `json:"visible"`
	MfrNumber           *string                `json:"mfr_number"`
}

type AddProfServiceInput struct {
	ItemNumber          string
	TypeID              int
	CptHcpcsCode        *string
	ScopeID             *int
	Description         *string
	Price               float64
	Cost                float64
	Sort1               *float64
	Sort2               *float64
	ReferringPhysician  bool
	Visible             bool
	MfrNumber           *string
}

type UpdateProfServiceInput struct {
	ItemNumber          *string
	CptHcpcsCode        *string
	TypeID              *int
	ScopeID             *int  // -1 = clear
	Description         *string
	Price               *float64
	Cost                *float64
	Sort1               *float64
	Sort2               *float64
	ReferringPhysician  *bool
	Visible             *bool
	MfrNumber           *string
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *Service) GetProfServiceTypes() ([]ProfServiceTypeResult, error) {
	var rows []service.ProfessionalServiceType
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ProfServiceTypeResult, len(rows))
	for i, r := range rows {
		result[i] = ProfServiceTypeResult{r.IDMedicalServiceType, r.Title}
	}
	return result, nil
}

func (s *Service) GetProfServiceScopes() ([]ProfServiceScopeResult, error) {
	var rows []service.ProfessionalServiceScope
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ProfServiceScopeResult, len(rows))
	for i, r := range rows {
		result[i] = ProfServiceScopeResult{r.IDProfessionalServiceScope, r.Title}
	}
	return result, nil
}

func (s *Service) GetProfServices(typeID *int) ([]ProfServiceListItem, error) {
	type row struct {
		IDProfessionalService int64
		ItemNumber            string
		CptHcpcsCode          *string
		ScopeTitle            *string
		TypeTitle             *string
		InvoiceDesc           *string
		Price                 float64
		Cost                  float64
	}

	q := s.db.Table("professional_service ps").
		Select("ps.id_professional_service, ps.item_number, ps.cpt_hcpcs_code, pss.title AS scope_title, pst.title AS type_title, ps.invoice_desc, ps.price, ps.cost").
		Joins("LEFT JOIN professional_service_scope pss ON pss.id_professional_service_scope = ps.professional_service_scope_id").
		Joins("LEFT JOIN professional_service_type pst ON pst.id_medical_service_type = ps.professional_service_type_id")

	if typeID != nil {
		q = q.Where("ps.professional_service_type_id = ?", *typeID)
	}

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]ProfServiceListItem, len(rows))
	for i, r := range rows {
		price := fmt.Sprintf("%.2f", r.Price)
		cost := fmt.Sprintf("%.2f", r.Cost)
		result[i] = ProfServiceListItem{
			ItemID:       r.IDProfessionalService,
			ItemName:     r.ItemNumber,
			CptHcpcsCode: r.CptHcpcsCode,
			Scope:        r.ScopeTitle,
			Type:         r.TypeTitle,
			Description:  r.InvoiceDesc,
			Price:        &price,
			Cost:         &cost,
			PbKey:        "Prof. service",
		}
	}
	return result, nil
}

func (s *Service) GetProfService(id int) (*ProfServiceDetail, error) {
	var svc service.ProfessionalService
	if err := s.db.
		Preload("Scope").
		Preload("Type").
		First(&svc, id).Error; err != nil {
		return nil, fmt.Errorf("professional service not found")
	}

	scope := map[string]interface{}{"scope_id": nil, "title": nil}
	if svc.Scope != nil {
		scope["scope_id"] = svc.Scope.IDProfessionalServiceScope
		scope["title"] = svc.Scope.Title
	}
	svcType := map[string]interface{}{"type_id": nil, "title": nil}
	if svc.Type != nil {
		svcType["type_id"] = svc.Type.IDMedicalServiceType
		svcType["title"] = svc.Type.Title
	}

	price := fmt.Sprintf("%.2f", svc.Price)
	cost := fmt.Sprintf("%.2f", svc.Cost)

	return &ProfServiceDetail{
		ServiceID:          svc.IDProfessionalService,
		ItemNumber:         svc.ItemNumber,
		CptHcpcsCode:       svc.CptHcpcsCode,
		Scope:              scope,
		Type:               svcType,
		Description:        svc.InvoiceDesc,
		Price:              &price,
		Cost:               &cost,
		Sort1:              svc.Sort1,
		Sort2:              svc.Sort2,
		ReferringPhysician: svc.ReferringPhysician,
		Visible:            svc.Visible,
		MfrNumber:          svc.MfrNumber,
	}, nil
}

func (s *Service) AddProfService(in AddProfServiceInput) (int64, error) {
	svc := service.ProfessionalService{
		ItemNumber:                in.ItemNumber,
		CptHcpcsCode:              in.CptHcpcsCode,
		ProfessionalServiceTypeID: &in.TypeID,
		ProfessionalServiceScopeID: in.ScopeID,
		InvoiceDesc:               in.Description,
		Price:                     in.Price,
		Cost:                      in.Cost,
		Sort1:                     in.Sort1,
		Sort2:                     in.Sort2,
		ReferringPhysician:        in.ReferringPhysician,
		Visible:                   in.Visible,
		MfrNumber:                 in.MfrNumber,
	}
	if err := s.db.Create(&svc).Error; err != nil {
		return 0, err
	}
	return svc.IDProfessionalService, nil
}

func (s *Service) UpdateProfService(id int, in UpdateProfServiceInput) error {
	var svc service.ProfessionalService
	if err := s.db.First(&svc, id).Error; err != nil {
		return fmt.Errorf("professional service not found")
	}

	updates := map[string]interface{}{}
	if in.ItemNumber != nil {
		updates["item_number"] = *in.ItemNumber
	}
	if in.CptHcpcsCode != nil {
		updates["cpt_hcpcs_code"] = *in.CptHcpcsCode
	}
	if in.TypeID != nil {
		updates["professional_service_type_id"] = *in.TypeID
	}
	if in.ScopeID != nil {
		if *in.ScopeID == -1 {
			updates["professional_service_scope_id"] = nil
		} else {
			updates["professional_service_scope_id"] = *in.ScopeID
		}
	}
	if in.Description != nil {
		updates["invoice_desc"] = *in.Description
	}
	if in.Price != nil {
		updates["price"] = *in.Price
	}
	if in.Cost != nil {
		updates["cost"] = *in.Cost
	}
	if in.Sort1 != nil {
		updates["sort1"] = *in.Sort1
	}
	if in.Sort2 != nil {
		updates["sort2"] = *in.Sort2
	}
	if in.ReferringPhysician != nil {
		updates["referring_physician"] = *in.ReferringPhysician
	}
	if in.Visible != nil {
		updates["visible"] = *in.Visible
	}
	if in.MfrNumber != nil {
		updates["mfr_number"] = *in.MfrNumber
	}

	if len(updates) > 0 {
		return s.db.Model(&svc).Updates(updates).Error
	}
	return nil
}

func (s *Service) DeleteProfService(id int) error {
	var svc service.ProfessionalService
	if err := s.db.First(&svc, id).Error; err != nil {
		return fmt.Errorf("professional service not found")
	}

	var countSale, countService int64
	s.db.Table("invoice_item_sale").Where("item_type = 'Prof. service' AND item_id = ?", id).Count(&countSale)
	s.db.Table("invoice_services_item").Where("professional_service_id = ?", id).Count(&countService)
	if countSale > 0 || countService > 0 {
		return fmt.Errorf("professional service is used in invoices")
	}

	return s.db.Delete(&svc).Error
}

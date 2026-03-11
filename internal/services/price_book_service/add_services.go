package price_book_service

import (
	"fmt"

	"sighthub-backend/internal/models/service"
)

// ─── Result types ─────────────────────────────────────────────────────────────

type AddTypeResult struct {
	TypeID   int    `json:"type_id"`
	TypeName string `json:"type_name"`
}

type AddServiceListItem struct {
	ItemID      int64   `json:"item_id"`
	ItemName    *string `json:"item_name"`
	ServiceType string  `json:"service_type"`
	Description string  `json:"description"`
	Price       string  `json:"price"`
	SrCost      *bool   `json:"sr_cost"`
	Tint        *bool   `json:"tint"`
	AR          *bool   `json:"ar"`
	UV          *bool   `json:"uv"`
	Drill       *bool   `json:"drill"`
	Send        *bool   `json:"send"`
	VCode       *string `json:"v_code"`
	PbKey       string  `json:"pb_key"`
}

type AddServiceInput struct {
	ItemNumber     string
	TypeID         int
	InvoiceDesc    string
	Price          float64
	CostPrice      float64
	SrCost         *bool
	UV             *bool
	AR             *bool
	Tint           *bool
	Drill          *bool
	Send           *bool
	ReportOmit     *bool
	InsVCode       *string
	ClassLevel     *string
	InsVCodeAdd    *string
	Sort1          *float64
	Sort2          *float64
	Visible        bool
	MfrNumber      *string
	Photochromatic *bool
	Polarized      *bool
	CanDrill       *bool
	HighIndex      *bool
	Digital        *bool
}

type UpdateAddServiceInput struct {
	ItemNumber     *string
	TypeID         *int
	InvoiceDesc    *string
	Price          *float64
	CostPrice      *float64
	SrCost         *bool
	UV             *bool
	AR             *bool
	Tint           *bool
	Drill          *bool
	Send           *bool
	ReportOmit     *bool
	InsVCode       *string
	ClassLevel     *string
	InsVCodeAdd    *string
	Sort1          *float64
	Sort2          *float64
	Visible        *bool
	MfrNumber      *string
	Photochromatic *bool
	Polarized      *bool
	CanDrill       *bool
	HighIndex      *bool
	Digital        *bool
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *Service) GetAddTypes() ([]AddTypeResult, error) {
	var rows []service.AdditionalServiceType
	if err := s.db.Distinct().Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]AddTypeResult, len(rows))
	for i, r := range rows {
		result[i] = AddTypeResult{r.IDAddServiceType, r.Title}
	}
	return result, nil
}

func (s *Service) GetAdditionalServices(typeID *int) ([]AddServiceListItem, error) {
	type row struct {
		IDAdditionalService int64
		ItemNumber          *string
		ServiceTypeTitle    string
		InvoiceDesc         string
		Price               float64
		SrCost              *bool
		Tint                *bool
		AR                  *bool
		UV                  *bool
		Drill               *bool
		Send                *bool
		InsVCode            *string
	}

	q := s.db.Table("additional_service ads").
		Select("ads.id_additional_service, ads.item_number, ast.title AS service_type_title, ads.invoice_desc, ads.price, ads.sr_cost, ads.tint, ads.ar, ads.uv, ads.drill, ads.send, ads.ins_v_code").
		Joins("JOIN add_service_type ast ON ast.id_add_service_type = ads.add_service_type_id")

	if typeID != nil {
		q = q.Where("ads.add_service_type_id = ?", *typeID)
	}

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]AddServiceListItem, len(rows))
	for i, r := range rows {
		result[i] = AddServiceListItem{
			ItemID:      r.IDAdditionalService,
			ItemName:    r.ItemNumber,
			ServiceType: r.ServiceTypeTitle,
			Description: r.InvoiceDesc,
			Price:       fmt.Sprintf("%.2f", r.Price),
			SrCost:      r.SrCost,
			Tint:        r.Tint,
			AR:          r.AR,
			UV:          r.UV,
			Drill:       r.Drill,
			Send:        r.Send,
			VCode:       r.InsVCode,
			PbKey:       "Add service",
		}
	}
	return result, nil
}

func (s *Service) GetAdditionalService(id int) (*service.AdditionalService, error) {
	var svc service.AdditionalService
	if err := s.db.
		Preload("AddServiceType").
		First(&svc, id).Error; err != nil {
		return nil, fmt.Errorf("additional service not found")
	}
	return &svc, nil
}

func (s *Service) AddAdditionalService(in AddServiceInput) (int64, error) {
	svc := service.AdditionalService{
		ItemNumber:     &in.ItemNumber,
		AddServiceTypeID: &in.TypeID,
		InvoiceDesc:    in.InvoiceDesc,
		Price:          in.Price,
		CostPrice:      in.CostPrice,
		SrCost:         in.SrCost,
		UV:             in.UV,
		AR:             in.AR,
		Tint:           in.Tint,
		Drill:          in.Drill,
		Send:           in.Send,
		ReportOmit:     in.ReportOmit,
		InsVCode:       in.InsVCode,
		ClassLevel:     in.ClassLevel,
		InsVCodeAdd:    in.InsVCodeAdd,
		Sort1:          in.Sort1,
		Sort2:          in.Sort2,
		Visible:        in.Visible,
		MfrNumber:      in.MfrNumber,
		Photochromatic: in.Photochromatic,
		Polarized:      in.Polarized,
		CanDrill:       in.CanDrill,
		HighIndex:      in.HighIndex,
		Digital:        in.Digital,
	}
	if err := s.db.Create(&svc).Error; err != nil {
		return 0, err
	}
	return svc.IDAdditionalService, nil
}

func (s *Service) UpdateAdditionalService(id int, in UpdateAddServiceInput) error {
	var svc service.AdditionalService
	if err := s.db.First(&svc, id).Error; err != nil {
		return fmt.Errorf("additional service not found")
	}

	updates := map[string]interface{}{}
	if in.ItemNumber != nil {
		updates["item_number"] = *in.ItemNumber
	}
	if in.TypeID != nil {
		updates["add_service_type_id"] = *in.TypeID
	}
	if in.InvoiceDesc != nil {
		updates["invoice_desc"] = *in.InvoiceDesc
	}
	if in.Price != nil {
		updates["price"] = *in.Price
	}
	if in.CostPrice != nil {
		updates["cost_price"] = *in.CostPrice
	}
	if in.SrCost != nil {
		updates["sr_cost"] = *in.SrCost
	}
	if in.UV != nil {
		updates["uv"] = *in.UV
	}
	if in.AR != nil {
		updates["ar"] = *in.AR
	}
	if in.Tint != nil {
		updates["tint"] = *in.Tint
	}
	if in.Drill != nil {
		updates["drill"] = *in.Drill
	}
	if in.Send != nil {
		updates["send"] = *in.Send
	}
	if in.ReportOmit != nil {
		updates["report_omit"] = *in.ReportOmit
	}
	if in.InsVCode != nil {
		updates["ins_v_code"] = *in.InsVCode
	}
	if in.ClassLevel != nil {
		updates["class_level"] = *in.ClassLevel
	}
	if in.InsVCodeAdd != nil {
		updates["ins_v_code_add"] = *in.InsVCodeAdd
	}
	if in.Sort1 != nil {
		updates["sort1"] = *in.Sort1
	}
	if in.Sort2 != nil {
		updates["sort2"] = *in.Sort2
	}
	if in.Visible != nil {
		updates["visible"] = *in.Visible
	}
	if in.MfrNumber != nil {
		updates["mfr_number"] = *in.MfrNumber
	}
	if in.Photochromatic != nil {
		updates["photochromatic"] = *in.Photochromatic
	}
	if in.Polarized != nil {
		updates["polarized"] = *in.Polarized
	}
	if in.CanDrill != nil {
		updates["can_drill"] = *in.CanDrill
	}
	if in.HighIndex != nil {
		updates["high_index"] = *in.HighIndex
	}
	if in.Digital != nil {
		updates["digital"] = *in.Digital
	}

	if len(updates) > 0 {
		return s.db.Model(&svc).Updates(updates).Error
	}
	return nil
}

func (s *Service) DeleteAdditionalService(id int) error {
	var svc service.AdditionalService
	if err := s.db.First(&svc, id).Error; err != nil {
		return fmt.Errorf("additional service not found")
	}

	var countSale, countInvoice int64
	s.db.Table("invoice_item_sale").Where("item_type = 'Add service' AND item_id = ?", id).Count(&countSale)
	s.db.Table("invoice_services_item").Where("additional_service_id = ?", id).Count(&countInvoice)
	if countSale > 0 || countInvoice > 0 {
		return fmt.Errorf("additional service is used in invoices")
	}

	return s.db.Delete(&svc).Error
}

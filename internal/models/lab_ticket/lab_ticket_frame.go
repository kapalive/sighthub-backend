// internal/models/lab_ticket/lab_ticket_frame.go
package lab_ticket

import (
	"fmt"

	"sighthub-backend/internal/models/frames"
)

// LabTicketFrame ↔ table: lab_ticket_frame
type LabTicketFrame struct {
	// старые поля
	IDLabTicketFrame    int64   `gorm:"column:id_lab_ticket_frame;primaryKey;autoIncrement" json:"id_lab_ticket_frame"`
	POF                 *string `gorm:"column:pof;type:text"                                 json:"pof,omitempty"`
	LabTicketStatus     int     `gorm:"column:lab_ticket_status;not null"                    json:"lab_ticket_status"`
	FrameTypeMaterialID *int    `gorm:"column:frame_type_material_id"                        json:"frame_type_material_id,omitempty"`
	AValue              *int    `gorm:"column:a_value"                                       json:"a_value,omitempty"`
	BValue              *int    `gorm:"column:b_value"                                       json:"b_value,omitempty"`
	EDValue             *int    `gorm:"column:ed_value"                                      json:"ed_value,omitempty"`
	CircValue           *int    `gorm:"column:circ_value"                                    json:"circ_value,omitempty"`

	// новые поля
	ModelTitleVariant *string `gorm:"column:model_title_variant;type:varchar(100)" json:"model_title_variant,omitempty"`
	MaterialsFrame    *string `gorm:"column:materials_frame;type:varchar(100)"     json:"materials_frame,omitempty"`
	MaterialsTemple   *string `gorm:"column:materials_temple;type:varchar(100)"    json:"materials_temple,omitempty"`
	Color             *string `gorm:"column:color;type:varchar(50)"                json:"color,omitempty"`
	SizeLensWidth     *string `gorm:"column:size_lens_width;type:varchar(3)"       json:"size_lens_width,omitempty"`
	SizeBridgeWidth   *string `gorm:"column:size_bridge_width;type:varchar(3)"     json:"size_bridge_width,omitempty"`
	SizeTempleLength  *string `gorm:"column:size_temple_length;type:varchar(4)"    json:"size_temple_length,omitempty"`

	Panto          *float64 `gorm:"column:panto"          json:"panto,omitempty"`
	WrapAngle      *float64 `gorm:"column:wrap_angle"     json:"wrap_angle,omitempty"`
	HeadEyeRatio   *float64 `gorm:"column:head_eye_ratio" json:"head_eye_ratio,omitempty"`
	StabilityCoeff *float64 `gorm:"column:stability_coeff" json:"stability_coeff,omitempty"`
	ERCX           *float64 `gorm:"column:erc_x"          json:"erc_x,omitempty"`
	ERCY           *float64 `gorm:"column:erc_y"          json:"erc_y,omitempty"`
	ERCZ           *float64 `gorm:"column:erc_z"          json:"erc_z,omitempty"`

	BC           *float64 `gorm:"column:bc"                                json:"bc,omitempty"`
	FrameShapeID *int     `gorm:"column:frame_shape_id"                    json:"frame_shape_id,omitempty"`
	FrameSource  *string  `gorm:"column:frame_source;type:varchar(100)"    json:"frame_source,omitempty"`
	ItemType     *string  `gorm:"column:item_type;type:varchar(50)"        json:"item_type,omitempty"`
	Status       *string  `gorm:"column:status;type:varchar(50)"           json:"status,omitempty"`

	DropShip bool    `gorm:"column:drop_ship;not null;default:false"   json:"drop_ship"`
	ShipTo   *string `gorm:"column:ship_to;type:varchar(100)"          json:"ship_to,omitempty"`

	FrameName        *string `gorm:"column:frame_name;type:varchar(255)"        json:"frame_name,omitempty"`
	BrandName        *string `gorm:"column:brand_name;type:varchar(255)"        json:"brand_name,omitempty"`
	VendorName       *string `gorm:"column:vendor_name;type:varchar(255)"       json:"vendor_name,omitempty"`
	ManufacturerName *string `gorm:"column:manufacturer_name;type:varchar(255)" json:"manufacturer_name,omitempty"`

	HeadCape  *string `gorm:"column:head_cape;type:varchar(6)" json:"head_cape,omitempty"`
	CorridorR *string `gorm:"column:corridor_r;type:varchar(6)" json:"corridor_r,omitempty"`
	CorridorL *string `gorm:"column:corridor_l;type:varchar(6)" json:"corridor_l,omitempty"`

	Read  *int `gorm:"column:read"   json:"read,omitempty"`
	VDFit *int `gorm:"column:vd_fit" json:"vd_fit,omitempty"`

	// связи (опционально прелоадить)
	FrameTypeMaterial *frames.FrameTypeMaterial `gorm:"foreignKey:FrameTypeMaterialID;references:IDFrameTypeMaterial" json:"-"`
	FrameShape        *frames.FrameShape        `gorm:"foreignKey:FrameShapeID;references:IDFrameShape"               json:"-"`
}

func (LabTicketFrame) TableName() string { return "lab_ticket_frame" }

func (l *LabTicketFrame) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_lab_ticket_frame":    l.IDLabTicketFrame,
		"pof":                    l.POF,
		"lab_ticket_status":      l.LabTicketStatus,
		"frame_type_material_id": l.FrameTypeMaterialID,
		"a_value":                l.AValue,
		"b_value":                l.BValue,
		"ed_value":               l.EDValue,
		"circ_value":             l.CircValue,

		"model_title_variant": l.ModelTitleVariant,
		"materials_frame":     l.MaterialsFrame,
		"materials_temple":    l.MaterialsTemple,
		"color":               l.Color,
		"size_lens_width":     l.SizeLensWidth,
		"size_bridge_width":   l.SizeBridgeWidth,
		"size_temple_length":  l.SizeTempleLength,

		"panto":           l.Panto,
		"wrap_angle":      l.WrapAngle,
		"head_eye_ratio":  l.HeadEyeRatio,
		"stability_coeff": l.StabilityCoeff,
		"erc_x":           l.ERCX,
		"erc_y":           l.ERCY,
		"erc_z":           l.ERCZ,

		"bc":             l.BC,
		"frame_shape_id": l.FrameShapeID,
		"frame_source":   l.FrameSource,
		"item_type":      l.ItemType,
		"status":         l.Status,

		"drop_ship": l.DropShip,
		"ship_to":   l.ShipTo,

		"frame_name":        l.FrameName,
		"brand_name":        l.BrandName,
		"vendor_name":       l.VendorName,
		"manufacturer_name": l.ManufacturerName,

		"head_cape":  l.HeadCape,
		"corridor_r": l.CorridorR,
		"corridor_l": l.CorridorL,

		"vd_fit": l.VDFit,
		"read":   l.Read,
	}

	// вложенные справочники — если прелоадили
	if l.FrameTypeMaterial != nil {
		m["frame_type_material"] = l.FrameTypeMaterial.ToMap()
	} else {
		m["frame_type_material"] = nil
	}
	if l.FrameShape != nil {
		m["frame_shape"] = l.FrameShape.ToMap()
	} else {
		m["frame_shape"] = nil
	}

	return m
}

func (l *LabTicketFrame) String() string {
	return fmt.Sprintf("<LabTicketFrame %d>", l.IDLabTicketFrame)
}

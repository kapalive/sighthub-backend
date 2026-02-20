// internal/models/lab_ticket/lab_ticket_lens.go
package lab_ticket

import (
	"fmt"

	lensmodel "sighthub-backend/internal/models/lenses"
)

// Postgres enum lab_ticket_lens_order_enum
type LabTicketLensOrder string

// LabTicketLens ↔ table: lab_ticket_lens
type LabTicketLens struct {
	IDLabTicketLens int64               `gorm:"column:id_lab_ticket_lens;primaryKey;autoIncrement"         json:"id_lab_ticket_lens"`
	LensStatus      *string             `gorm:"column:lens_status;type:varchar(255)"                        json:"lens_status,omitempty"`
	LensOrder       *LabTicketLensOrder `gorm:"column:lens_order;type:lab_ticket_lens_order_enum"           json:"lens_order,omitempty"`

	LensTypesID           *int    `gorm:"column:lens_types_id"                 json:"lens_types_id,omitempty"`
	LensesMaterialsID     *int    `gorm:"column:lenses_materials_id"           json:"lenses_materials_id,omitempty"`
	EdgeThickness         *string `gorm:"column:edge_thickness"                json:"edge_thickness,omitempty"`
	CenterThickness       *string `gorm:"column:center_thickness"              json:"center_thickness,omitempty"`
	LensSafetyThicknessID *int    `gorm:"column:lens_safety_thickness_id"      json:"lens_safety_thickness_id,omitempty"`
	LensEdgeID            *int    `gorm:"column:lens_edge_id"                  json:"lens_edge_id,omitempty"`
	LensTintColorID       *int    `gorm:"column:lens_tint_color_id"            json:"lens_tint_color_id,omitempty"`
	LensTypeColor         *string `gorm:"column:lens_type_color"               json:"lens_type_color,omitempty"`
	TintPercent           *int    `gorm:"column:tint_percent"                  json:"tint_percent,omitempty"`
	FadeColor             *string `gorm:"column:fade_color"                    json:"fade_color,omitempty"`
	SolidColor            *string `gorm:"column:solid_color"                   json:"solid_color,omitempty"`
	LensSampleColorID     *int    `gorm:"column:lens_sample_color_id"          json:"lens_sample_color_id,omitempty"`
	NotesColor            *string `gorm:"column:notes_color"                   json:"notes_color,omitempty"`

	// Optional preload relations (импортируем пакет lenses прямо)
	LensType            *lensmodel.LensType            `gorm:"foreignKey:LensTypesID;references:IDLensType"                   json:"-"`
	LensesMaterial      *lensmodel.LensesMaterial      `gorm:"foreignKey:LensesMaterialsID;references:IDLensesMaterials"      json:"-"`
	LensSafetyThickness *lensmodel.LensSafetyThickness `gorm:"foreignKey:LensSafetyThicknessID;references:IDLensSafetyThickness" json:"-"`
	LensEdge            *lensmodel.LensEdge            `gorm:"foreignKey:LensEdgeID;references:IDLensEdge"                    json:"-"`
	LensTintColor       *lensmodel.LensTintColor       `gorm:"foreignKey:LensTintColorID;references:IDLensTintColor"          json:"-"`
	LensSampleColor     *lensmodel.LensSampleColor     `gorm:"foreignKey:LensSampleColorID;references:IDLensSampleColor"      json:"-"`
}

func (LabTicketLens) TableName() string { return "lab_ticket_lens" }

func (l *LabTicketLens) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lab_ticket_lens":       l.IDLabTicketLens,
		"lens_status":              l.LensStatus,
		"lens_order":               l.LensOrder,
		"lens_types_id":            l.LensTypesID,
		"lenses_materials_id":      l.LensesMaterialsID,
		"edge_thickness":           l.EdgeThickness,
		"center_thickness":         l.CenterThickness,
		"lens_safety_thickness_id": l.LensSafetyThicknessID,
		"lens_edge_id":             l.LensEdgeID,
		"lens_tint_color_id":       l.LensTintColorID,
		"lens_type_color":          l.LensTypeColor,
		"tint_percent":             l.TintPercent,
		"fade_color":               l.FadeColor,
		"solid_color":              l.SolidColor,
		"lens_sample_color_id":     l.LensSampleColorID,
		"notes_color":              l.NotesColor,
	}
}

func (l *LabTicketLens) String() string {
	return fmt.Sprintf("<LabTicketLens %d>", l.IDLabTicketLens)
}

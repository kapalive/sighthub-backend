package lenses

import "fmt"

type LensEdge struct {
	IDLensEdge   int    `gorm:"column:id_lens_edge;primaryKey" json:"id_lens_edge"`
	LensEdgeName string `gorm:"column:lens_edge_name;type:varchar(255);not null" json:"lens_edge_name"`
	Description  string `gorm:"column:description;type:text" json:"description"`
}

func (LensEdge) TableName() string {
	return "lens_edge"
}

func (l *LensEdge) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lens_edge":   l.IDLensEdge,
		"lens_edge_name": l.LensEdgeName,
		"description":    l.Description,
	}
}

func (l *LensEdge) String() string {
	return fmt.Sprintf("<LensEdge %s>", l.LensEdgeName)
}

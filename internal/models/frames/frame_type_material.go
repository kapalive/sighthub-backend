// internal/models/frames/frame_type_material.go
package frames

import "fmt"

type FrameTypeMaterial struct {
	IDFrameTypeMaterial int    `gorm:"column:id_frame_type_material;primaryKey" json:"id_frame_type_material"`
	Material            string `gorm:"column:material;type:varchar(255);not null" json:"material"`
}

func (FrameTypeMaterial) TableName() string { return "frame_type_material" }

func (f *FrameTypeMaterial) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"frame_type_material_id": f.IDFrameTypeMaterial,
		"material":               f.Material,
	}
}

func (f *FrameTypeMaterial) String() string {
	return fmt.Sprintf("<FrameTypeMaterial %s>", f.Material)
}

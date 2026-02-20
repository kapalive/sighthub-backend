// internal/models/frames/frame_shape.go
package frames

import "fmt"

type FrameShape struct {
	IDFrameShape    int     `gorm:"column:id_frame_shape;primaryKey"                              json:"id_frame_shape"`
	TitleFrameShape string  `gorm:"column:title_frame_shape;type:varchar(100);not null;unique"    json:"title_frame_shape"`
	Description     *string `gorm:"column:description;type:text"                                  json:"description,omitempty"`
}

func (FrameShape) TableName() string { return "frame_shape" }

func (f *FrameShape) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_frame_shape":    f.IDFrameShape,
		"title_frame_shape": f.TitleFrameShape,
	}
	if f.Description != nil {
		m["description"] = *f.Description
	} else {
		m["description"] = nil
	}
	return m
}

func (f *FrameShape) String() string {
	return fmt.Sprintf("<FrameShape %s>", f.TitleFrameShape)
}

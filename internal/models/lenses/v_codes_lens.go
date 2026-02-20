// internal/models/lenses/v_codes_lens.go
package lenses

import "fmt"

type VCodesLens struct {
	IDVCodesLens int    `gorm:"column:id_v_codes_lens;primaryKey" json:"id_v_codes_lens"`
	Code         string `gorm:"column:code;type:varchar(50);not null;uniqueIndex:ux_v_codes_lens_code" json:"code"`
	Description  string `gorm:"column:description;type:text" json:"description"`
}

func (VCodesLens) TableName() string {
	return "v_codes_lens"
}

func (v *VCodesLens) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_v_codes_lens": v.IDVCodesLens,
		"code":            v.Code,
		"description":     v.Description,
	}
}

func (v *VCodesLens) String() string {
	return fmt.Sprintf("<VCodesLens %s>", v.Code)
}

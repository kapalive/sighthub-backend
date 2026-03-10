package external_sle

type VisualFields struct {
	IDVisualFields   int64   `gorm:"column:id_visual_fields;primaryKey;autoIncrement" json:"id_visual_fields"`
	SuperoTemporoOd  *string `gorm:"column:supero_temporo_od;size:255" json:"supero_temporo_od"`
	SuperoTemporoOs  *string `gorm:"column:supero_temporo_os;size:255" json:"supero_temporo_os"`
	SuperoNasalOd    *string `gorm:"column:supero_nasal_od;size:255" json:"supero_nasal_od"`
	SuperoNasalOs    *string `gorm:"column:supero_nasal_os;size:255" json:"supero_nasal_os"`
	InferoTemporalOd *string `gorm:"column:infero_temporal_od;size:255" json:"infero_temporal_od"`
	InferoTemporalOs *string `gorm:"column:infero_temporal_os;size:255" json:"infero_temporal_os"`
	InferoNasalOd    *string `gorm:"column:infero_nasal_od;size:255" json:"infero_nasal_od"`
	InferoNasalOs    *string `gorm:"column:infero_nasal_os;size:255" json:"infero_nasal_os"`
	Instrument       *string `gorm:"column:instrument;size:255" json:"instrument"`
	Test             *string `gorm:"column:test;size:255" json:"test"`
	Reason           *string `gorm:"column:reason;size:255" json:"reason"`
	Result           *string `gorm:"column:result;size:255" json:"result"`
	Recommendations  *string `gorm:"column:recommendations;size:255" json:"recommendations"`
	Comments         *string `gorm:"column:comments;type:text" json:"comments"`
}

func (VisualFields) TableName() string { return "visual_fields" }

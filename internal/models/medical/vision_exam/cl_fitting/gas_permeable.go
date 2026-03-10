package cl_fitting

// GasPermeable ↔ table: gas_permeable
type GasPermeable struct {
	IDGasPermeable int64  `gorm:"column:id_gas_permeable;primaryKey;autoIncrement" json:"id_gas_permeable"`
	LabDesignID    *int64 `gorm:"column:lab_design_id"                             json:"lab_design_id,omitempty"`
	DrDesignID     *int64 `gorm:"column:dr_design_id"                              json:"dr_design_id,omitempty"`

	LabDesign *LabDesign `gorm:"foreignKey:LabDesignID;references:IDLabDesign" json:"-"`
	DrDesign  *DrDesign  `gorm:"foreignKey:DrDesignID;references:IDDrDesign"   json:"-"`
}
func (GasPermeable) TableName() string { return "gas_permeable" }
func (g *GasPermeable) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_gas_permeable": g.IDGasPermeable,
		"lab_design_id": g.LabDesignID, "dr_design_id": g.DrDesignID,
	}
	if g.LabDesign != nil { m["lab_design"] = g.LabDesign.ToMap() }
	if g.DrDesign != nil  { m["dr_design"]  = g.DrDesign.ToMap() }
	return m
}

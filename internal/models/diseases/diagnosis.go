// internal/models/diseases/diagnosis.go
package diseases

// Diagnosis ↔ table: diagnosis (ICD-10 level 5)
type Diagnosis struct {
	IDDiagnosis    int64  `gorm:"column:id_diagnosis;primaryKey;autoIncrement"  json:"id_diagnosis"`
	Level4ID       int64  `gorm:"column:level_4_id;not null"                    json:"level_4_id"`
	Code           string `gorm:"column:code;type:varchar(10);not null"         json:"code"`
	TitleDiagnosis string `gorm:"column:title_diagnosis;type:varchar(255);not null" json:"title_diagnosis"`
	FullName       string `gorm:"column:full_name;type:varchar(266);not null"   json:"full_name"`

	Level4 *SubgroupDiseaseLevel4 `gorm:"foreignKey:Level4ID;references:IDLevel4" json:"-"`
}

func (Diagnosis) TableName() string { return "diagnosis" }

func (d *Diagnosis) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_diagnosis":    d.IDDiagnosis,
		"level_4_id":      d.Level4ID,
		"code":            d.Code,
		"title_diagnosis": d.TitleDiagnosis,
		"full_name":       d.FullName,
	}
}

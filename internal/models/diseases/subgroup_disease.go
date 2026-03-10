// internal/models/diseases/subgroup_disease.go
package diseases

// SubgroupDisease ↔ table: subgroup_disease (ICD-10 subgroup)
type SubgroupDisease struct {
	IDSubgroupDisease               int64  `gorm:"column:id_subgroup_disease;primaryKey;autoIncrement"       json:"id_subgroup_disease"`
	GroupDiseaseIDGroupDisease      int64  `gorm:"column:group_disease_id_group_disease;not null"            json:"group_disease_id_group_disease"`
	Code                            string `gorm:"column:code;type:varchar(10);not null"                     json:"code"`
	Title                           string `gorm:"column:title;type:varchar(255);not null"                   json:"title"`

	Group *GroupDisease `gorm:"foreignKey:GroupDiseaseIDGroupDisease;references:IDGroupDisease" json:"-"`
}

func (SubgroupDisease) TableName() string { return "subgroup_disease" }

func (s *SubgroupDisease) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_subgroup_disease":           s.IDSubgroupDisease,
		"group_disease_id_group_disease": s.GroupDiseaseIDGroupDisease,
		"code":                           s.Code,
		"title":                          s.Title,
	}
}

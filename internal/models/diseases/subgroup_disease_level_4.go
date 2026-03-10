// internal/models/diseases/subgroup_disease_level_4.go
package diseases

// SubgroupDiseaseLevel4 ↔ table: subgroup_disease_level_4 (ICD-10 level 4)
type SubgroupDiseaseLevel4 struct {
	IDLevel4                        int64  `gorm:"column:id_level_4;primaryKey;autoIncrement"                    json:"id_level_4"`
	SubgroupDiseaseIDSubgroupDisease int64  `gorm:"column:subgroup_disease_id_subgroup_disease;not null"          json:"subgroup_disease_id_subgroup_disease"`
	Code                            string `gorm:"column:code;type:varchar(10);not null;uniqueIndex"             json:"code"`
	TitleLevel4                     string `gorm:"column:title_level_4;type:varchar(255);not null"               json:"title_level_4"`

	Subgroup *SubgroupDisease `gorm:"foreignKey:SubgroupDiseaseIDSubgroupDisease;references:IDSubgroupDisease" json:"-"`
}

func (SubgroupDiseaseLevel4) TableName() string { return "subgroup_disease_level_4" }

func (s *SubgroupDiseaseLevel4) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_level_4":                          s.IDLevel4,
		"subgroup_disease_id_subgroup_disease": s.SubgroupDiseaseIDSubgroupDisease,
		"code":                                 s.Code,
		"title_level_4":                        s.TitleLevel4,
	}
}

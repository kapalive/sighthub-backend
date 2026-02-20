package insurance

import "fmt"

type InsuranceCoverageType struct {
	IDInsuranceCoverageType int    `gorm:"column:id_insurance_coverage_type;primaryKey"                               json:"id_insurance_coverage_type"`
	CoverageName            string `gorm:"column:coverage_name;type:varchar(255);not null;uniqueIndex:uniq_coverage_name" json:"coverage_name"`
}

func (InsuranceCoverageType) TableName() string { return "insurance_coverage_types" }

func (t *InsuranceCoverageType) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_insurance_coverage_type": t.IDInsuranceCoverageType,
		"coverage_name":              t.CoverageName,
	}
}

func (t *InsuranceCoverageType) String() string {
	return fmt.Sprintf("<InsuranceCoverageType %s>", t.CoverageName)
}

// internal/models/general/ethnicity.go
package general

import "fmt"

type Ethnicity struct {
	IDEthnicity   int    `gorm:"column:id_ethnicity;primaryKey"                  json:"id_ethnicity"`
	EthnicityName string `gorm:"column:ethnicity_name;type:varchar(50);not null" json:"ethnicity_name"`
}

func (Ethnicity) TableName() string { return "ethnicity" }

func (e *Ethnicity) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_ethnicity":   e.IDEthnicity,
		"ethnicity_name": e.EthnicityName,
	}
}

func (e *Ethnicity) String() string {
	return fmt.Sprintf("<Ethnicity %s>", e.EthnicityName)
}

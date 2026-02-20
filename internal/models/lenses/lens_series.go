// internal/models/lenses/lens_series.go
package lenses

import "fmt"

type LensSeries struct {
	IDLensSeries int    `gorm:"column:id_lens_series;primaryKey" json:"id_lens_series"`
	SeriesName   string `gorm:"column:series_name;type:varchar(255);not null" json:"series_name"`
	Description  string `gorm:"column:description;type:text" json:"description"`
}

func (LensSeries) TableName() string {
	return "lens_series"
}

func (l *LensSeries) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lens_series": l.IDLensSeries,
		"series_name":    l.SeriesName,
		"description":    l.Description,
	}
}

func (l *LensSeries) String() string {
	return fmt.Sprintf("<LensSeries %s>", l.SeriesName)
}

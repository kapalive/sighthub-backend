// internal/models/medical/vision_exam/chief_complaint/chief_complaint_hpi_eye.go
package chief_complaint

// ChiefComplaintHPIEye ↔ table: chief_complaint_hpi_eye
type ChiefComplaintHPIEye struct {
	IDChiefComplaintHPIEye int64   `gorm:"column:id_chief_complaint_hpi_eye;primaryKey;autoIncrement" json:"id_chief_complaint_hpi_eye"`
	NoteChiefComplaint     *string `gorm:"column:note_chief_complaint;type:text"                      json:"note_chief_complaint,omitempty"`
	Location               *string `gorm:"column:location;type:text"                                  json:"location,omitempty"`
	Quality                *string `gorm:"column:quality;type:text"                                   json:"quality,omitempty"`
	Severity               *string `gorm:"column:severity;type:text"                                  json:"severity,omitempty"`
	Duration               *string `gorm:"column:duration;type:text"                                  json:"duration,omitempty"`
	Timing                 *string `gorm:"column:timing;type:text"                                    json:"timing,omitempty"`
	Context                *string `gorm:"column:context;type:text"                                   json:"context,omitempty"`
	Factors                *string `gorm:"column:factors;type:text"                                   json:"factors,omitempty"`
	Symptoms               *string `gorm:"column:symptoms;type:text"                                  json:"symptoms,omitempty"`
}

func (ChiefComplaintHPIEye) TableName() string { return "chief_complaint_hpi_eye" }

func (c *ChiefComplaintHPIEye) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_chief_complaint_hpi_eye": c.IDChiefComplaintHPIEye,
		"note_chief_complaint":       c.NoteChiefComplaint,
		"location":                   c.Location,
		"quality":                    c.Quality,
		"severity":                   c.Severity,
		"duration":                   c.Duration,
		"timing":                     c.Timing,
		"context":                    c.Context,
		"factors":                    c.Factors,
		"symptoms":                   c.Symptoms,
	}
}

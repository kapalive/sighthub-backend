// internal/models/medical/vision_exam/chief_complaint/secondary_complaint_hpi_eye.go
package chief_complaint

// SecondaryComplaintHPIEye ↔ table: secondary_complaint_hpi_eye
type SecondaryComplaintHPIEye struct {
	IDSecondaryComplaintHPIEye int64   `gorm:"column:id_secondary_complaint_hpi_eye;primaryKey;autoIncrement" json:"id_secondary_complaint_hpi_eye"`
	NoteSecondaryComplaint     *string `gorm:"column:note_secondary_complaint;type:text"                      json:"note_secondary_complaint,omitempty"`
	Location                   *string `gorm:"column:location;type:text"                                      json:"location,omitempty"`
	Quality                    *string `gorm:"column:quality;type:text"                                       json:"quality,omitempty"`
	Severity                   *string `gorm:"column:severity;type:text"                                      json:"severity,omitempty"`
	Duration                   *string `gorm:"column:duration;type:text"                                      json:"duration,omitempty"`
	Timing                     *string `gorm:"column:timing;type:text"                                        json:"timing,omitempty"`
	Context                    *string `gorm:"column:context;type:text"                                       json:"context,omitempty"`
	Factors                    *string `gorm:"column:factors;type:text"                                       json:"factors,omitempty"`
	Symptoms                   *string `gorm:"column:symptoms;type:text"                                      json:"symptoms,omitempty"`
}

func (SecondaryComplaintHPIEye) TableName() string { return "secondary_complaint_hpi_eye" }

func (s *SecondaryComplaintHPIEye) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_secondary_complaint_hpi_eye": s.IDSecondaryComplaintHPIEye,
		"note_secondary_complaint":       s.NoteSecondaryComplaint,
		"location":                       s.Location,
		"quality":                        s.Quality,
		"severity":                       s.Severity,
		"duration":                       s.Duration,
		"timing":                         s.Timing,
		"context":                        s.Context,
		"factors":                        s.Factors,
		"symptoms":                       s.Symptoms,
	}
}

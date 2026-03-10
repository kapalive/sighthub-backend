// internal/models/general/planing_communication.go
package general

import "time"

// PlaningCommunication ⇄ planing_communication
type PlaningCommunication struct {
	IDPlaningCommunication int64      `gorm:"column:id_planing_communication;primaryKey;autoIncrement" json:"id_planing_communication"`
	PatientID              int64      `gorm:"column:patient_id;not null;index"                         json:"patient_id"`
	CommunicationTypeID    int        `gorm:"column:communication_type_id;not null"                    json:"communication_type_id"`
	Reason                 *string    `gorm:"column:reason;type:varchar(255)"                          json:"reason,omitempty"`
	Date                   time.Time  `gorm:"column:date;type:timestamptz;not null"                    json:"date"`
	LocationID             *int64     `gorm:"column:location_id"                                       json:"location_id,omitempty"`
	Note                   *string    `gorm:"column:note;type:text"                                    json:"note,omitempty"`
	SourceTable            *string    `gorm:"column:source_table;type:varchar(50);index"               json:"source_table,omitempty"`
	SourceID               *int64     `gorm:"column:source_id;index"                                   json:"source_id,omitempty"`
}

func (PlaningCommunication) TableName() string { return "planing_communication" }

func (p *PlaningCommunication) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_planing_communication": p.IDPlaningCommunication,
		"patient_id":               p.PatientID,
		"communication_type_id":    p.CommunicationTypeID,
		"reason":                   p.Reason,
		"note":                     p.Note,
		"date":                     p.Date.Format(time.RFC3339),
		"location_id":              p.LocationID,
		"source_table":             p.SourceTable,
		"source_id":                p.SourceID,
	}
}

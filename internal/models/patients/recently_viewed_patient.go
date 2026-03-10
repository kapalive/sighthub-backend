package patients

import "time"

// RecentlyViewedPatient ⇄ table: recently_viewed_patient
type RecentlyViewedPatient struct {
	IDViewed       int64      `gorm:"column:id_viewed;primaryKey;autoIncrement"                                json:"id_viewed"`
	LocationID     int        `gorm:"column:location_id;not null;index"                                        json:"location_id"`
	PatientID      int64      `gorm:"column:patient_id;not null;index;uniqueIndex:uix_recent_view_location_patient" json:"patient_id"`
	DatetimeViewed *time.Time `gorm:"column:datetime_viewed;type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"datetime_viewed,omitempty"`

	Patient *Patient `gorm:"foreignKey:PatientID;references:IDPatient" json:"-"`
}

func (RecentlyViewedPatient) TableName() string { return "recently_viewed_patient" }
